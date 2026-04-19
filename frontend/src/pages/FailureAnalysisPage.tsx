import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Checkbox, message, Space, Table, Tabs, Typography, Upload } from 'antd';
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
  listFailureFeatureFacts,
  listStorageTopServerRates,
  exportStorageTopServers,
  listPackageFailureRates,
  listPackageModelFailureRates
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  FaultAnalysisResult,
  FaultYearAnalysisRow,
  FailureFeatureFact,
  StorageTopServerRate,
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
  const [excludeOverWarranty, setExcludeOverWarranty] = useState(false);

  const [overallRates, setOverallRates] = useState<FaultAnalysisResult['overall_rates']>([]);
  const [featureFacts, setFeatureFacts] = useState<FailureFeatureFact[]>([]);
  const [warmStorageTopRates, setWarmStorageTopRates] = useState<StorageTopServerRate[]>([]);
  const [hotStorageTopRates, setHotStorageTopRates] = useState<StorageTopServerRate[]>([]);
  const [yearFaultRows, setYearFaultRows] = useState<FaultYearAnalysisRow[]>([]);
  const [fm, setFm] = useState<ModelFailureRate[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);

  async function reloadAll() {
    try {
      const [s0, s1, s2, s3, s4, s5, s6] = await Promise.all([
        listOverallFailureRates(),
        listModelFailureRates(),
        listPackageFailureRates(),
        listPackageModelFailureRates(),
        listFailureFeatureFacts(),
        listStorageTopServerRates('warm_storage'),
        listStorageTopServerRates('hot_storage')
      ]);
      setOverallRates(ensureApiOk(s0).data.list || []);
      setFm(ensureApiOk(s1).data.list);
      setFp(ensureApiOk(s2).data.list);
      setFpm(ensureApiOk(s3).data.list);
      setFeatureFacts(ensureApiOk(s4).data.list || []);
      setWarmStorageTopRates(ensureApiOk(s5).data.list || []);
      setHotStorageTopRates(ensureApiOk(s6).data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载失败'));
    }
  }

  useEffect(() => {
    reloadAll();
  }, []);

  const phase1SceneAgeSummary = useMemo(() => {
    const sceneOrder = ['storage', 'non_storage', 'compute', 'warm_storage', 'hot_storage', 'gpu'];
    const m = new Map<string, { scene_group: string; age_bucket: number; denominator_weighted: number; fault_count: number }>();
    featureFacts
      .filter((x) => x.scope === 'all')
      .forEach((x) => {
        const key = `${x.scene_group}|${x.age_bucket}`;
        const old = m.get(key) || { scene_group: x.scene_group, age_bucket: x.age_bucket, denominator_weighted: 0, fault_count: 0 };
        old.denominator_weighted += Number(x.denominator_weighted || 0);
        old.fault_count += Number(x.fault_count || 0);
        m.set(key, old);
      });
    const out = Array.from(m.values()).map((x) => ({
      ...x,
      fault_rate: x.denominator_weighted > 0 ? x.fault_count / x.denominator_weighted : 0
    }));
    out.sort((a, b) => {
      const ai = sceneOrder.indexOf(a.scene_group);
      const bi = sceneOrder.indexOf(b.scene_group);
      if (ai !== bi) return (ai < 0 ? 999 : ai) - (bi < 0 ? 999 : bi);
      return a.age_bucket - b.age_bucket;
    });
    return out;
  }, [featureFacts]);

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
            const resp = ensureApiOk(await analyzeFaultRates(file, { excludeOverWarranty }));
            setAnalysisResult(resp.data);
            setOverallRates(resp.data.overall_rates || []);
            setFeatureFacts(resp.data.failure_feature_facts || []);
            setYearFaultRows(resp.data.year_fault_analysis_rows || []);
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

  const storageTopColumns = [
    { title: 'SN', dataIndex: 'sn' },
    { title: '厂商', dataIndex: 'manufacturer' },
    { title: '型号', dataIndex: 'model' },
    { title: '配置类型', dataIndex: 'config_type' },
    { title: '环境', dataIndex: 'environment' },
    { title: '数据盘数', dataIndex: 'data_disk_count', render: (v: number) => formatInt(v) },
    { title: '单盘容量(TB)', dataIndex: 'single_disk_capacity_tb', render: (v: number) => formatFloat(v) },
    { title: '单台总容量(TB)', dataIndex: 'total_capacity_tb', render: (v: number) => formatFloat(v) },
    { title: '最近1年故障次数', dataIndex: 'fault_count', render: (v: number) => formatInt(v) },
    { title: '分母(1+盘数)', dataIndex: 'denominator', render: (v: number) => formatFloat(v) },
    { title: '故障率', dataIndex: 'fault_rate', render: (v: number) => formatPercent(v) }
  ];

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
        title="整体故障率（按存储/非存储 + 全周期/超5年）"
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
            { title: '超5年故障率', dataIndex: 'over_warranty_failure_rate', render: (v: number) => formatPercent(v) },
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
                <Space direction="vertical" size={8}>
                  <Checkbox checked={excludeOverWarranty} onChange={(e) => setExcludeOverWarranty(e.target.checked)}>
                    排除过保服务器（按保修日期）
                  </Checkbox>
                  <Text type="secondary">
                    上传包含故障字段的清单后，系统会结合服务器管理表 + 主机套餐配置表自动重算型号/套餐/套餐型号故障率。
                  </Text>
                </Space>
              </Card>
            )
          },
          {
            key: 'phase1',
            label: '高级特性Phase1',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Card title="分场景 x 机龄汇总（跨记录年求和）">
                  <Text type="secondary">用于透明展示分子分母：按场景、机龄桶把所有记录年求和后再算故障率。</Text>
                  <Table rowKey={(r) => `${r.scene_group}-${r.age_bucket}`} dataSource={phase1SceneAgeSummary} pagination={{ pageSize: 12 }} columns={[
                    { title: '场景', dataIndex: 'scene_group', render: (v: string) => sceneLabel(v) },
                    { title: '机龄桶', dataIndex: 'age_bucket', render: (v: number) => v >= 9 ? '8+' : `${v}` },
                    { title: '分母(加权求和)', dataIndex: 'denominator_weighted', render: (v: number) => formatFloat(v) },
                    { title: '分子(故障次数求和)', dataIndex: 'fault_count', render: (v: number) => formatInt(v) },
                    { title: '故障率', dataIndex: 'fault_rate', render: (v: number) => formatPercent(v) }
                  ]} />
                </Card>

                <Card title="记录年 x 机龄桶（Phase1 明细）">
                  <Table rowKey={(r) => `${r.record_year_index}-${r.scope}-${r.scene_group}-${r.age_bucket}`} dataSource={featureFacts} pagination={{ pageSize: 12 }} columns={[
                    { title: '记录年', dataIndex: 'record_year_index', render: (_: number, r: FailureFeatureFact) => `Y${r.record_year_index} (${r.record_year_start}~${r.record_year_end})` },
                    { title: '范围', dataIndex: 'scope', render: (v: string) => scopeLabel(v) },
                    { title: '场景', dataIndex: 'scene_group', render: (v: string) => sceneLabel(v) },
                    { title: '机龄桶', dataIndex: 'age_bucket', render: (v: number) => v >= 9 ? '8+' : `${v}` },
                    { title: '分母(加权)', dataIndex: 'denominator_weighted', render: (v: number) => formatFloat(v) },
                    { title: '故障次数', dataIndex: 'fault_count', render: (v: number) => formatInt(v) },
                    { title: '故障率', dataIndex: 'fault_rate', render: (v: number) => formatPercent(v) }
                  ]} />
                </Card>
              </Space>
            )
          },
          {
            key: 'warm-storage-top100',
            label: '温存储故障TOP100',
            children: (
              <Card
                title="最近1年温存储故障服务器TOP100"
                extra={<Space>
                  <Button onClick={() => exportStorageTopServers('warm_storage', 'xlsx')}>下载温存储详细数据</Button>
                </Space>}
              >
                <Text type="secondary">仅统计温存储服务器；公式：最近1年故障次数 / (1 + 数据盘数量)。下载文件含全部温存储清单与保修截止日期。</Text>
                <Table rowKey="sn" dataSource={warmStorageTopRates} pagination={{ pageSize: 20 }} columns={storageTopColumns} />
              </Card>
            )
          },
          {
            key: 'hot-storage-top100',
            label: '热存储故障TOP100',
            children: (
              <Card
                title="最近1年热存储故障服务器TOP100"
                extra={<Space>
                  <Button onClick={() => exportStorageTopServers('hot_storage', 'xlsx')}>下载热存储详细数据</Button>
                </Space>}
              >
                <Text type="secondary">仅统计热存储服务器；公式：最近1年故障次数 / (1 + 数据盘数量)。下载文件含全部热存储清单与保修截止日期。</Text>
                <Table rowKey="sn" dataSource={hotStorageTopRates} pagination={{ pageSize: 20 }} columns={storageTopColumns} />
              </Card>
            )
          },
          {
            key: 'year-fault-analysis',
            label: `${new Date().getFullYear()}年故障率分析`,
            children: (
              <Card title={`${new Date().getFullYear()}年故障率命中分析（基于导入故障清单）`}>
                <Text type="secondary">用于核对“整体存储/整体非存储故障数”与样本清单差异。命中=参与当年故障率计算；未命中可查看备注原因。</Text>
                <Table
                  rowKey={(r) => `${r.row_no}-${r.sn || ''}-${r.created_at || ''}`}
                  dataSource={yearFaultRows}
                  pagination={{ pageSize: 20 }}
                  columns={[
                    { title: '行号', dataIndex: 'row_no', width: 80 },
                    { title: 'SN', dataIndex: 'sn', width: 180, render: (v?: string) => v || '-' },
                    { title: '创建时间', dataIndex: 'created_at', width: 180, render: (v?: string) => v || '-' },
                    { title: '范围', dataIndex: 'scope', width: 100, render: (v?: string) => scopeLabel(v) },
                    { title: '分类', dataIndex: 'segment', width: 100, render: (v?: string) => (v === 'storage' ? '存储' : v === 'non_storage' ? '非存储' : '-') },
                    { title: '命中', dataIndex: 'matched', width: 90, render: (v: boolean) => (v ? '是' : '否') },
                    { title: '备注', dataIndex: 'remark', width: 380, render: (v?: string) => v || '-' }
                  ]}
                  scroll={{ x: 1100 }}
                />
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

function sceneLabel(v?: string) {
  if (v === 'storage') return '存储';
  if (v === 'non_storage') return '非存储';
  if (v === 'compute') return '计算';
  if (v === 'warm_storage') return '温存储';
  if (v === 'hot_storage') return '热存储';
  if (v === 'gpu') return 'GPU';
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
