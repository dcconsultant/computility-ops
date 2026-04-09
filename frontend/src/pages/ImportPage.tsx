import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { Alert, Button, Card, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import type { UploadProps } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import {
  importHostPackages,
  importModelFailureRates,
  importPackageFailureRates,
  importPackageModelFailureRates,
  importServers,
  importSpecialRules,
  listHostPackages,
  listModelFailureRates,
  listPackageFailureRates,
  listPackageModelFailureRates,
  listServers,
  listSpecialRules
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  HostPackageConfig,
  ImportResult,
  ModelFailureRate,
  PackageFailureRate,
  PackageModelFailureRate,
  ServerItem,
  SpecialRule
} from '../types';

const { Text } = Typography;

type DataKey = 'servers' | 'packages' | 'special' | 'failure_model' | 'failure_package' | 'failure_pkg_model';

const titles: Record<DataKey, string> = {
  servers: '服务器管理',
  packages: '主机套餐配置',
  special: '特殊名单',
  failure_model: '型号故障率',
  failure_package: '套餐故障率',
  failure_pkg_model: '套餐型号故障率'
};

export default function ImportPage() {
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [uploading, setUploading] = useState<DataKey | null>(null);

  const [servers, setServers] = useState<ServerItem[]>([]);
  const [packages, setPackages] = useState<HostPackageConfig[]>([]);
  const [special, setSpecial] = useState<SpecialRule[]>([]);
  const [fm, setFm] = useState<ModelFailureRate[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);

  async function reloadAll() {
    try {
      const [s1, s2, s3, s4, s5, s6] = await Promise.all([
        listServers(),
        listHostPackages(),
        listSpecialRules(),
        listModelFailureRates(),
        listPackageFailureRates(),
        listPackageModelFailureRates()
      ]);
      setServers(ensureApiOk(s1).data.list);
      setPackages(ensureApiOk(s2).data.list);
      setSpecial(ensureApiOk(s3).data.list);
      setFm(ensureApiOk(s4).data.list);
      setFp(ensureApiOk(s5).data.list);
      setFpm(ensureApiOk(s6).data.list);
    } catch (e) {
      message.error(parseApiError(e, '加载失败'));
    }
  }

  useEffect(() => {
    reloadAll();
  }, []);

  function makeUploadProps(kind: DataKey): UploadProps {
    const importer = {
      servers: importServers,
      packages: importHostPackages,
      special: importSpecialRules,
      failure_model: importModelFailureRates,
      failure_package: importPackageFailureRates,
      failure_pkg_model: importPackageModelFailureRates
    }[kind];

    return {
      maxCount: 1,
      showUploadList: true,
      accept: '.xlsx',
      customRequest: async (options) => {
        const file = options.file as File;
        setUploading(kind);
        try {
          const resp = ensureApiOk(await importer(file));
          setImportResult(resp.data);
          message.success(`${titles[kind]}导入完成：成功 ${resp.data.success} 条`);
          await reloadAll();
          options.onSuccess?.({}, new XMLHttpRequest());
        } catch (e) {
          message.error(parseApiError(e, '导入失败'));
          options.onError?.(new Error('import failed'));
        } finally {
          setUploading(null);
        }
      }
    };
  }

  const tableCard = (title: string, kind: DataKey, table: ReactNode, helper: string) => (
    <Card title={title} extra={<Upload {...makeUploadProps(kind)}><Button icon={<UploadOutlined />} loading={uploading === kind}>上传并导入</Button></Upload>}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Text type="secondary">{helper}</Text>
        {table}
      </Space>
    </Card>
  );

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {importResult && (
        <Alert
          showIcon
          type={importResult.failed > 0 ? 'warning' : 'success'}
          message={`总计 ${importResult.total}，成功 ${importResult.success}，失败 ${importResult.failed}`}
          description={importResult.errors.length ? importResult.errors.slice(0, 5).map((e) => `第${e.row}行: ${e.reason}`).join('；') : undefined}
        />
      )}

      <Tabs
        items={[
          {
            key: 'servers',
            label: '服务器管理',
            children: tableCard(
              '服务器管理表',
              'servers',
              <Table rowKey="sn" dataSource={servers} pagination={{ pageSize: 10 }} columns={[
                { title: 'SN', dataIndex: 'sn' },
                { title: '制造商', dataIndex: 'manufacturer' },
                { title: '型号', dataIndex: 'model' },
                { title: 'PSA', dataIndex: 'psa' },
                { title: '机房', dataIndex: 'idc' },
                { title: '环境', dataIndex: 'environment' },
                { title: '配置类型', dataIndex: 'config_type' },
                { title: '保修结束日期', dataIndex: 'warranty_end_date' },
                { title: '投产日期', dataIndex: 'launch_date' }
              ]} />,
              '字段：SN、制造商、型号、PSA、机房、环境、配置类型、保修结束日期、投产日期'
            )
          },
          {
            key: 'packages',
            label: '主机套餐配置',
            children: tableCard(
              '主机套餐配置表',
              'packages',
              <Table rowKey="config_type" dataSource={packages} pagination={{ pageSize: 10 }} columns={[
                { title: '配置类型', dataIndex: 'config_type' },
                { title: '场景大类', dataIndex: 'scene_category' },
                { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores' },
                { title: '磁盘类型', dataIndex: 'disk_type' },
                { title: '存储容量(TB)', dataIndex: 'storage_capacity_tb' },
                { title: '架构标准化系数', dataIndex: 'arch_standardized_factor' }
              ]} />,
              '服务器管理表通过配置类型关联此表；排名按 PSA × 架构标准化系数。'
            )
          },
          {
            key: 'special',
            label: '特殊名单',
            children: tableCard(
              '特殊名单',
              'special',
              <Table rowKey="sn" dataSource={special} pagination={{ pageSize: 10 }} columns={[
                { title: 'SN', dataIndex: 'sn' },
                { title: '制造商', dataIndex: 'manufacturer' },
                { title: '型号', dataIndex: 'model' },
                { title: '套餐', dataIndex: 'package_type' },
                { title: '策略', dataIndex: 'policy' }
              ]} />,
              '策略列请填：加白/加黑（或 whitelist/blacklist）。加白强制续保，加黑强制不续保。'
            )
          },
          {
            key: 'failure',
            label: '故障率入口',
            children: (
              <Tabs
                items={[
                  {
                    key: 'fm',
                    label: '型号故障率',
                    children: tableCard('型号故障率表', 'failure_model', <Table rowKey={(r) => `${r.manufacturer}-${r.model}`} dataSource={fm} pagination={{ pageSize: 10 }} columns={[
                      { title: '厂商', dataIndex: 'manufacturer' },
                      { title: '型号', dataIndex: 'model' },
                      { title: '故障率', dataIndex: 'failure_rate' }
                    ]} />, '本版本暂不参与续保算法，仅提供统一入口维护。')
                  },
                  {
                    key: 'fp',
                    label: '套餐故障率',
                    children: tableCard('套餐故障率表', 'failure_package', <Table rowKey="config_type" dataSource={fp} pagination={{ pageSize: 10 }} columns={[
                      { title: '配置类型', dataIndex: 'config_type' },
                      { title: '故障率', dataIndex: 'failure_rate' }
                    ]} />, '本版本暂不参与续保算法，仅提供统一入口维护。')
                  },
                  {
                    key: 'fpm',
                    label: '套餐型号故障率',
                    children: tableCard('套餐型号故障率表', 'failure_pkg_model', <Table rowKey={(r) => `${r.config_type}-${r.manufacturer}-${r.model}`} dataSource={fpm} pagination={{ pageSize: 10 }} columns={[
                      { title: '套餐', dataIndex: 'config_type' },
                      { title: '厂商', dataIndex: 'manufacturer' },
                      { title: '型号', dataIndex: 'model' },
                      { title: '故障率', dataIndex: 'failure_rate' }
                    ]} />, '本版本暂不参与续保算法，仅提供统一入口维护。')
                  }
                ]}
              />
            )
          }
        ]}
      />
    </Space>
  );
}
