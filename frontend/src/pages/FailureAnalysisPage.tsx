import { useEffect, useState } from 'react';
import { Alert, Button, Card, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import { useNavigate } from 'react-router-dom';
import type { UploadProps } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import {
  analyzeFaultRates,
  importModelFailureRates,
  importPackageFailureRates,
  importPackageModelFailureRates,
  listModelFailureRates,
  listOverallFailureRates,
  listPackageFailureRates,
  listPackageModelFailureRates
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  FaultAnalysisResult,
  ModelFailureRate,
  PackageFailureRate,
  PackageModelFailureRate
} from '../types';

const { Text } = Typography;

type DataKey = 'failure_model' | 'failure_package' | 'failure_pkg_model' | 'fault_analysis';

export default function FailureAnalysisPage() {
  const navigate = useNavigate();
  const [analysisResult, setAnalysisResult] = useState<FaultAnalysisResult | null>(null);
  const [uploading, setUploading] = useState<DataKey | null>(null);

  const [overallRates, setOverallRates] = useState<FaultAnalysisResult['overall_rates']>([]);
  const [fm, setFm] = useState<ModelFailureRate[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);

  async function reloadAll() {
    try {
      const [s0, s1, s2, s3] = await Promise.all([
        listOverallFailureRates(),
        listModelFailureRates(),
        listPackageFailureRates(),
        listPackageModelFailureRates()
      ]);
      setOverallRates(ensureApiOk(s0).data.list || []);
      setFm(ensureApiOk(s1).data.list);
      setFp(ensureApiOk(s2).data.list);
      setFpm(ensureApiOk(s3).data.list);
    } catch (e) {
      message.error(parseApiError(e, '加载失败'));
    }
  }

  useEffect(() => {
    reloadAll();
  }, []);

  function makeUploadProps(kind: DataKey): UploadProps {
    const importer = {
      failure_model: importModelFailureRates,
      failure_package: importPackageFailureRates,
      failure_pkg_model: importPackageModelFailureRates
    }[kind as Exclude<DataKey, 'fault_analysis'>];

    return {
      maxCount: 1,
      showUploadList: true,
      accept: '.xlsx',
      customRequest: async (options) => {
        const file = options.file as File;
        setUploading(kind);
        try {
          if (kind === 'fault_analysis') {
            const resp = ensureApiOk(await analyzeFaultRates(file));
            setAnalysisResult(resp.data);
            setOverallRates(resp.data.overall_rates || []);
            await reloadAll();
            message.success(`故障分析完成：命中故障 ${resp.data.matched_fault_rows}/${resp.data.total_fault_rows} 条`);
          } else {
            const resp = ensureApiOk(await importer(file));
            message.success(`导入完成：成功 ${resp.data.success} 条`);
            await reloadAll();
          }
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

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {analysisResult && (
        <Alert
          showIcon
          type="success"
          message={`故障分析完成：故障清单 ${analysisResult.total_fault_rows} 行，命中真实故障 ${analysisResult.matched_fault_rows} 行`}
          description={`生成故障率：型号 ${analysisResult.generated_model_rates} 条、套餐 ${analysisResult.generated_package_rates} 条、套餐型号 ${analysisResult.generated_package_model_rates} 条`}
        />
      )}

      <Card
        title="整体故障率（按存储/非存储 + 全周期/过保）"
        extra={<Button type="primary" onClick={() => navigate('/failure/dashboard')}>打开分析看板</Button>}
      >
        <Table
          rowKey="segment"
          pagination={false}
          dataSource={overallRates || []}
          columns={[
            { title: '分组', dataIndex: 'period', render: (_: string, r: any) => periodLabel(r.period, r.year) },
            { title: '环境范围', dataIndex: 'scope', render: (v: string) => scopeLabel(v) },
            { title: '分类', dataIndex: 'segment', render: (v: string) => (v === 'storage' ? '存储' : '非存储') },
            { title: '全周期故障率', dataIndex: 'full_cycle_failure_rate', render: (v: number) => formatPercent(v) },
            { title: '过保故障率', dataIndex: 'over_warranty_failure_rate', render: (v: number) => formatPercent(v) },
            { title: '故障数', dataIndex: 'fault_count', render: (v: number) => formatInt(v) },
            { title: '过保故障数', dataIndex: 'over_warranty_fault_count', render: (v: number) => formatInt(v) },
            { title: '全周期台年', dataIndex: 'server_years', render: (v: number) => formatFloat(v) },
            { title: '过保台年', dataIndex: 'over_warranty_years', render: (v: number) => formatFloat(v) }
          ]}
        />
      </Card>

      <Tabs
        items={[
          {
            key: 'fm',
            label: '型号故障率',
            children: (
              <Card title="型号故障率表" extra={<Upload {...makeUploadProps('failure_model')}><Button icon={<UploadOutlined />} loading={uploading === 'failure_model'}>上传并导入</Button></Upload>}>
                <Text type="secondary">故障率为年化值；过保故障率按投产满5年至今区间统计，仅供参考。</Text>
                <Table rowKey={(r) => `${r.manufacturer}-${r.model}`} dataSource={fm} pagination={{ pageSize: 10 }} columns={[
                  { title: '厂商', dataIndex: 'manufacturer' },
                  { title: '服务器型号', dataIndex: 'model' },
                  { title: '年化故障率', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) },
                  { title: '过保故障率(参考)', dataIndex: 'over_warranty_failure_rate', render: (v: number) => formatPercent(v) }
                ]} />
              </Card>
            )
          },
          {
            key: 'fp',
            label: '套餐故障率',
            children: (
              <Card title="套餐故障率表" extra={<Upload {...makeUploadProps('failure_package')}><Button icon={<UploadOutlined />} loading={uploading === 'failure_package'}>上传并导入</Button></Upload>}>
                <Text type="secondary">故障率为年化值；过保故障率按投产满5年至今区间统计，仅供参考。</Text>
                <Table rowKey="config_type" dataSource={fp} pagination={{ pageSize: 10 }} columns={[
                  { title: '配置类型', dataIndex: 'config_type' },
                  { title: '年化故障率', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) },
                  { title: '过保故障率(参考)', dataIndex: 'over_warranty_failure_rate', render: (v: number) => formatPercent(v) }
                ]} />
              </Card>
            )
          },
          {
            key: 'fpm',
            label: '套餐型号故障率',
            children: (
              <Card title="套餐型号故障率表" extra={<Upload {...makeUploadProps('failure_pkg_model')}><Button icon={<UploadOutlined />} loading={uploading === 'failure_pkg_model'}>上传并导入</Button></Upload>}>
                <Text type="secondary">故障率为年化值；过保故障率按投产满5年至今区间统计，仅供参考。</Text>
                <Table rowKey={(r) => `${r.config_type}-${r.manufacturer}-${r.model}`} dataSource={fpm} pagination={{ pageSize: 10 }} columns={[
                  { title: '套餐', dataIndex: 'config_type' },
                  { title: '厂商', dataIndex: 'manufacturer' },
                  { title: '服务器型号', dataIndex: 'model' },
                  { title: '年化故障率', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) },
                  { title: '过保故障率(参考)', dataIndex: 'over_warranty_failure_rate', render: (v: number) => formatPercent(v) }
                ]} />
              </Card>
            )
          },
          {
            key: 'fa',
            label: '故障清单分析',
            children: (
              <Card title="上传故障清单并自动分析" extra={<Upload {...makeUploadProps('fault_analysis')}><Button icon={<UploadOutlined />} loading={uploading === 'fault_analysis'}>上传并分析</Button></Upload>}>
                <Text type="secondary">
                  上传包含故障字段的清单后，系统会结合服务器管理表 + 主机套餐配置表自动重算型号/套餐/套餐型号故障率。
                </Text>
              </Card>
            )
          }
        ]}
      />
    </Space>
  );
}

function periodLabel(period?: string, year?: number) {
  if (period === 'history') return '历史平均故障率';
  if (period === 'year') return `${year || new Date().getFullYear()}年故障率`;
  return period || '-';
}

function scopeLabel(v?: string) {
  if (v === 'all') return '整体';
  if (v === 'product') return '生产';
  if (v === 'devtest') return '开测';
  return v || '-';
}

function formatPercent(v?: number) {
  const n = Number(v || 0);
  return `${(n * 100).toFixed(2)}%`;
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
