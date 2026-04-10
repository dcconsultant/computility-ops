import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { Alert, Button, Card, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import type { UploadProps } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import {
  importHostPackages,
  importServers,
  listHostPackages,
  listServers
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  HostPackageConfig,
  ImportResult,
  ServerItem
} from '../types';

const { Text } = Typography;

type DataKey = 'servers' | 'packages';

const titles: Record<DataKey, string> = {
  servers: '服务器管理',
  packages: '主机套餐配置'
};

export default function ImportPage() {
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [uploading, setUploading] = useState<DataKey | null>(null);

  const [servers, setServers] = useState<ServerItem[]>([]);
  const [packages, setPackages] = useState<HostPackageConfig[]>([]);

  async function reloadAll() {
    try {
      const [s1, s2] = await Promise.all([
        listServers(),
        listHostPackages()
      ]);
      setServers(ensureApiOk(s1).data.list);
      setPackages(ensureApiOk(s2).data.list);
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
      packages: importHostPackages
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
                { title: '服务器型号', dataIndex: 'model' },
                { title: 'PSA', dataIndex: 'psa' },
                { title: '机房', dataIndex: 'idc' },
                { title: '环境', dataIndex: 'environment' },
                { title: '配置类型', dataIndex: 'config_type' },
                { title: '保修结束日期', dataIndex: 'warranty_end_date' },
                { title: '投产日期', dataIndex: 'launch_date' }
              ]} />,
              '字段：SN、制造商、服务器型号、PSA、机房、环境、配置类型、保修结束日期、投产日期'
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
                { title: 'GPU卡数', dataIndex: 'gpu_card_count' },
                { title: '数据盘类型', dataIndex: 'data_disk_type' },
                { title: '数据盘数量', dataIndex: 'data_disk_count' },
                { title: '存储容量(TB)', dataIndex: 'storage_capacity_tb' },
                { title: '服务器价值分', dataIndex: 'server_value_score' },
                { title: '架构标准化系数', dataIndex: 'arch_standardized_factor' }
              ]} />,
              '服务器管理表通过配置类型关联此表；需维护服务器价值分（PSA非数字时基准）与GPU卡数（GPU汇总统计依赖）。'
            )
          }
        ]}
      />
    </Space>
  );
}
