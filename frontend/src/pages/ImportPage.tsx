import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { ReactNode } from 'react';
import { Alert, Button, Card, Col, Input, InputNumber, Modal, Popconfirm, Row, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import type { UploadProps } from 'antd';
import { DownloadOutlined, UploadOutlined } from '@ant-design/icons';
import {
  createCabinetConfig,
  deleteCabinetConfig,
  exportServerPackageAnomalies,
  exportHostPackageTemplate,
  exportValueScorePerformanceParamsTemplate,
  getCabinetUtilization,
  getValueScoreCabinetBaseline,
  getValueScoreCostParams,
  updateValueScoreCostParams,
  calculateValueScoreTCO,
  calculateValueScorePerformance,
  exportValueScoreTCO,
  importHostPackages,
  importServers,
  importCabinetConfigs,
  importValueScoreUnifiedParams,
  previewValueScorePerformanceParams,
  listValueScorePerformanceParams,
  exportCabinetTemplate,
  getRenewalSettings,
  listCabinetConfigs,
  listHostPackages,
  listMetaModels,
  listMetaRecords,
  listServers,
  updateCabinetConfig,
  updateCabinetUtilization
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  CabinetConfig,
  HostPackageConfig,
  ImportResult,
  MetaRecord,
  RenewalPlanSettings,
  ServerItem,
  ValueScoreCabinetBaseline,
  ValueScoreCostParams,
  ValueScorePerformanceCalcItem,
  ValueScoreTCOResult
} from '../types';

const { Text } = Typography;

type DataKey = 'servers' | 'packages' | 'cabinet' | 'value_score_setup' | 'assets';

const titles: Record<DataKey, string> = {
  servers: '服务器管理',
  packages: '主机套餐配置',
  cabinet: '机柜配置管理',
  value_score_setup: '价值分管理',
  assets: '资产分析'
};

const COMPUTE_STANDARD_PERFORMANCE_THRESHOLD = 870;

export default function ImportPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [uploading, setUploading] = useState<DataKey | null>(null);

  const [servers, setServers] = useState<ServerItem[]>([]);
  const [packages, setPackages] = useState<HostPackageConfig[]>([]);
  const [serverKeyword, setServerKeyword] = useState('');
  const [packageKeyword, setPackageKeyword] = useState('');
  const [cabinetUtilization, setCabinetUtilization] = useState<number>(1);
  const [cabinetRows, setCabinetRows] = useState<CabinetConfig[]>([]);
  const [cabinetModalOpen, setCabinetModalOpen] = useState(false);
  const [editingCabinet, setEditingCabinet] = useState<CabinetConfig | null>(null);
  const [cabinetForm, setCabinetForm] = useState({ idc: '', rated_power_kw: 0, monthly_rent: 0 });
  const [cabinetBaseline, setCabinetBaseline] = useState<ValueScoreCabinetBaseline | null>(null);
  const [costParams, setCostParams] = useState<ValueScoreCostParams>({
    depreciation_months: 60,
    network_device_share_cny: 0,
    server_renewal_fee_cny: 0,
    cabinet_utilization: 1,
    rated_power_kw: 0,
    monthly_rent_cny: 0
  });
  const [performancePreview, setPerformancePreview] = useState<any>(null);
  const [performanceResult, setPerformanceResult] = useState<{ items: ValueScorePerformanceCalcItem[]; alert_count: number; note?: string } | null>(null);
  const [tcoResult, setTcoResult] = useState<ValueScoreTCOResult | null>(null);
  const [tcoLoading, setTcoLoading] = useState(false);
  const [performanceLoading, setPerformanceLoading] = useState(false);
  const [valueConfigVisible, setValueConfigVisible] = useState(false);
  const [valueSceneTab, setValueSceneTab] = useState<'compute' | 'warm_storage' | 'hot_storage' | 'gpu'>('compute');
  const [valueConfigKeyword, setValueConfigKeyword] = useState('');
  const [assetMetaServers, setAssetMetaServers] = useState<MetaRecord[]>([]);
  const [assetMetaRacks, setAssetMetaRacks] = useState<MetaRecord[]>([]);
  const [assetMetaConfigs, setAssetMetaConfigs] = useState<MetaRecord[]>([]);
  const [renewalSettings, setRenewalSettings] = useState<RenewalPlanSettings | null>(null);

  async function reloadAll() {
    try {
      const [s1, s2, s3, s4, s5, s6, s7, s8, s9, s10] = await Promise.all([
        listServers(),
        listHostPackages(),
        getCabinetUtilization(),
        listCabinetConfigs(),
        getValueScoreCabinetBaseline(),
        getValueScoreCostParams(),
        listValueScorePerformanceParams(),
        calculateValueScorePerformance(),
        getRenewalSettings(),
        loadAssetMetaRecords()
      ]);
      setServers((ensureApiOk(s1) as any).data.list || []);
      setPackages((ensureApiOk(s2) as any).data.list || []);
      setCabinetUtilization((ensureApiOk(s3) as any).data.utilization || 1);
      setCabinetRows((ensureApiOk(s4) as any).data.list || []);
      setCabinetBaseline((ensureApiOk(s5) as any).data);
      setCostParams((ensureApiOk(s6) as any).data);
      setPerformancePreview((ensureApiOk(s7) as any).data);
      setPerformanceResult((ensureApiOk(s8) as any).data);
      setRenewalSettings((ensureApiOk(s9) as any).data);
      setAssetMetaServers(s10.servers);
      setAssetMetaRacks(s10.racks);
      setAssetMetaConfigs(s10.configs);
      try {
        const tco = ensureApiOk(await calculateValueScoreTCO());
        setTcoResult(tco.data);
      } catch {
        setTcoResult(null);
      }
    } catch (e) {
      message.error(parseApiError(e, '加载失败'));
    }
  }

  async function loadAssetMetaRecords() {
    const modelsResp = ensureApiOk(await listMetaModels());
    const models = modelsResp.data.list || [];
    const serverModel = models.find((m) => m.model_code === 'server');
    const rackModel = models.find((m) => m.model_code === 'rack');
    const configModel = models.find((m) => m.model_code === 'config_type');
    const [serverRows, rackRows, configRows] = await Promise.all([
      serverModel ? listMetaRecords(serverModel.id) : Promise.resolve(null),
      rackModel ? listMetaRecords(rackModel.id) : Promise.resolve(null),
      configModel ? listMetaRecords(configModel.id) : Promise.resolve(null)
    ]);
    return {
      servers: serverRows ? ensureApiOk(serverRows).data.list || [] : [],
      racks: rackRows ? ensureApiOk(rackRows).data.list || [] : [],
      configs: configRows ? ensureApiOk(configRows).data.list || [] : []
    };
  }

  useEffect(() => {
    reloadAll();
  }, []);

  const filteredServers = useMemo(() => {
    const q = serverKeyword.trim().toLowerCase();
    if (!q) return servers;
    return servers.filter((x) => [
      x.sn,
      x.manufacturer,
      x.model,
      x.psa,
      x.detailed_config,
      x.idc,
      x.environment,
      x.config_type,
      x.config_type_standardized,
      x.warranty_end_date,
      x.launch_date
    ].some((v) => (v || '').toString().toLowerCase().includes(q)));
  }, [servers, serverKeyword]);

  const filteredPackages = useMemo(() => {
    const q = packageKeyword.trim().toLowerCase();
    if (!q) return packages;
    return packages.filter((x) => [
      x.config_type,
      x.scene_category,
      x.cpu_logical_cores,
      x.gpu_card_count,
      x.data_disk_type,
      x.data_disk_count,
      x.storage_capacity_tb,
      x.power_watts,
      x.release_year,
      x.memory_capacity_gb,
      x.arch_standardized_factor
    ].some((v) => String(v ?? '').toLowerCase().includes(q)));
  }, [packages, packageKeyword]);

  const assetAnalysis = useMemo(() => buildAssetAnalysis({
    legacyServers: servers,
    metaServers: assetMetaServers,
    metaRacks: assetMetaRacks,
    metaConfigs: assetMetaConfigs,
    valuePackages: packages,
    performanceItems: performanceResult?.items || [],
    idleStoppedPSAs: renewalSettings?.idle_stopped_psas || []
  }), [servers, assetMetaServers, assetMetaRacks, assetMetaConfigs, packages, performanceResult, renewalSettings]);

  function onExportUnmatchedRackCSV() {
    const rows = assetAnalysis.unmatchedRackRows || [];
    if (!rows.length) {
      message.warning('暂无机柜未匹配服务器清单可下载');
      return;
    }
    const headers = ['服务器SN', '机柜', '配置类型', 'PSA', '机房', '未匹配原因'];
    const csv = [headers, ...rows.map((r) => [r.sn, r.rack, r.configType, r.psa, r.idc, r.reason])]
      .map((row) => row.map(csvCell).join(','))
      .join('\n');
    downloadCSV(`unmatched-rack-servers-${formatDateTime(new Date())}.csv`, csv);
  }

  const section = searchParams.get('section') || '';
  const requestedTab = (searchParams.get('tab') as DataKey | null) || null;

  const visibleTabKeys: DataKey[] = section === 'value-score'
    ? ['value_score_setup']
    : section === 'resource-analysis'
      ? ['assets']
      : section === 'test-zone'
        ? ['servers', 'packages', 'cabinet']
        : ['servers', 'packages', 'cabinet', 'value_score_setup', 'assets'];

  const activeTab = requestedTab && visibleTabKeys.includes(requestedTab)
    ? requestedTab
    : visibleTabKeys[0];

  const mergedScoreRows = useMemo(() => {
    const perfItems = (performanceResult?.items || []) as any[];
    const tcoItems = (tcoResult?.items || []) as any[];
    const tcoMap = new Map<string, any>();
    for (const t of tcoItems) tcoMap.set(t.config_type, t);
    const packageMap = new Map<string, HostPackageConfig>();
    for (const p of packages) packageMap.set(p.config_type, p);

    const classifyScene = (pkg?: HostPackageConfig): 'compute' | 'warm_storage' | 'hot_storage' | 'gpu' => {
      const appCategory = String((pkg as any)?.application_category || '').trim();
      if (appCategory === 'GPU') return 'gpu';
      if (appCategory === '温存储') return 'warm_storage';
      if (appCategory === '热存储') return 'hot_storage';
      if (appCategory === '计算') return 'compute';

      // 兼容历史数据兜底
      const scene = String(pkg?.scene_category || '').toLowerCase();
      if (Number(pkg?.gpu_card_count || 0) > 0 || scene.includes('gpu')) return 'gpu';
      if (scene.includes('warm') || scene.includes('温')) return 'warm_storage';
      if (scene.includes('hot') || scene.includes('热')) return 'hot_storage';
      return 'compute';
    };

    return perfItems.map((p) => {
      const t = tcoMap.get(p.config_type) || {};
      const pkg = packageMap.get(p.config_type);
      const sceneType = classifyScene(pkg);
      const totalTCO = Number(t.total_tco_monthly || 0);
      const availableCores = Number(p.available_cores || 0);
      const overallRatio = Number(p.overall_performance_ratio || 0);
      const storageTB = Number(pkg?.storage_capacity_tb || 0);
      const gpuCount = Number(pkg?.gpu_card_count || 0);

      let unitTCO = 0;
      let convertedUnitTCO = 0;
      let appliedOverallRatio = overallRatio;
      if (sceneType === 'gpu') {
        unitTCO = gpuCount > 0 ? totalTCO / gpuCount : 0;
        appliedOverallRatio = 1;
        convertedUnitTCO = unitTCO;
      } else if (sceneType === 'warm_storage' || sceneType === 'hot_storage') {
        unitTCO = storageTB > 0 ? totalTCO / storageTB : 0;
        appliedOverallRatio = 1;
        convertedUnitTCO = unitTCO;
      } else {
        unitTCO = availableCores > 0 ? totalTCO / availableCores : 0;
        convertedUnitTCO = unitTCO * overallRatio;
      }

      const valueScoreV1 = sceneType === 'compute' ? (convertedUnitTCO / 30) : unitTCO;

      return {
        ...p,
        ...t,
        scene_type: sceneType,
        scene_category: (pkg as any)?.application_category || pkg?.scene_category,
        capacity_storage_tb: storageTB,
        count_gpu: gpuCount,
        unit_tco: Number.isFinite(unitTCO) ? unitTCO : 0,
        overall_performance_ratio: Number.isFinite(appliedOverallRatio) ? appliedOverallRatio : 0,
        converted_unit_tco: Number.isFinite(convertedUnitTCO) ? convertedUnitTCO : 0,
        value_score_v1: Number.isFinite(valueScoreV1) ? valueScoreV1 : 0
      };
    });
  }, [performanceResult, tcoResult, packages]);

  const sceneScoreRows = useMemo(() => ({
    compute: mergedScoreRows.filter((r: any) => r.scene_type === 'compute'),
    warm_storage: mergedScoreRows.filter((r: any) => r.scene_type === 'warm_storage'),
    hot_storage: mergedScoreRows.filter((r: any) => r.scene_type === 'hot_storage'),
    gpu: mergedScoreRows.filter((r: any) => r.scene_type === 'gpu')
  }), [mergedScoreRows]);

  const filteredSceneScoreRows = useMemo(() => {
    const q = valueConfigKeyword.trim().toLowerCase();
    if (!q) return sceneScoreRows;
    const filterByConfig = (rows: any[]) => rows.filter((r) => String(r.config_type || '').toLowerCase().includes(q));
    return {
      compute: filterByConfig(sceneScoreRows.compute),
      warm_storage: filterByConfig(sceneScoreRows.warm_storage),
      hot_storage: filterByConfig(sceneScoreRows.hot_storage),
      gpu: filterByConfig(sceneScoreRows.gpu)
    };
  }, [sceneScoreRows, valueConfigKeyword]);

  const textSorter = (key: string) => (a: any, b: any) => String(a[key] || '').localeCompare(String(b[key] || ''), 'zh-Hans-CN');
  const numberSorter = (key: string) => (a: any, b: any) => Number(a[key] || 0) - Number(b[key] || 0);

  function makeUploadProps(kind: 'servers' | 'packages'): UploadProps {
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

  const tableCard = (title: string, kind: 'servers' | 'packages', table: ReactNode, helper: string, extra?: ReactNode) => (
    <Card
      title={title}
      extra={extra || <Upload {...makeUploadProps(kind)}><Button icon={<UploadOutlined />} loading={uploading === kind}>上传并导入</Button></Upload>}
    >
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
        activeKey={activeTab}
        onChange={(key) => {
          const next = new URLSearchParams(searchParams);
          next.set('tab', key);
          setSearchParams(next, { replace: true });
        }}
        items={[
          {
            key: 'servers',
            label: '服务器管理',
            children: tableCard(
              '服务器管理表',
              'servers',
              <Space direction="vertical" style={{ width: '100%' }}>
                <Input allowClear placeholder="搜索服务器（SN/型号/PSA/环境/配置类型等）" value={serverKeyword} onChange={(e) => setServerKeyword(e.target.value)} />
                <Table rowKey="sn" dataSource={filteredServers} pagination={withTotalPagination(10)} columns={[
                  { title: 'SN', dataIndex: 'sn' },
                  { title: '制造商', dataIndex: 'manufacturer' },
                  { title: '服务器型号', dataIndex: 'model' },
                  { title: '详细配置', dataIndex: 'detailed_config', width: 220, ellipsis: true },
                  { title: 'PSA', dataIndex: 'psa', render: (v: string) => formatMaybeNumber(v) },
                  { title: '机房', dataIndex: 'idc' },
                  { title: '环境', dataIndex: 'environment' },
                  { title: '配置类型', dataIndex: 'config_type' },
                  { title: '配置类型标准化', dataIndex: 'config_type_standardized' },
                  { title: '保修结束日期', dataIndex: 'warranty_end_date' },
                  { title: '投产日期', dataIndex: 'launch_date' }
                ]} />
              </Space>,
              '字段：SN、制造商、服务器型号、详细配置、PSA、机房、环境、配置类型、配置类型标准化、保修结束日期、投产日期',
              <Space>
                <Button onClick={() => exportServerPackageAnomalies('xlsx')}>下载套餐标准化异常清单</Button>
                <Upload {...makeUploadProps('servers')}><Button icon={<UploadOutlined />} loading={uploading === 'servers'}>上传并导入</Button></Upload>
              </Space>
            )
          },
          {
            key: 'packages',
            label: '主机套餐配置',
            children: tableCard(
              '主机套餐配置表',
              'packages',
              <Space direction="vertical" style={{ width: '100%' }}>
                <Input allowClear placeholder="搜索套餐（配置类型/场景/核数/卡数/存储等）" value={packageKeyword} onChange={(e) => setPackageKeyword(e.target.value)} />
                <Table rowKey="config_type" dataSource={filteredPackages} pagination={withTotalPagination(10)} columns={[
                  { title: '配置类型', dataIndex: 'config_type' },
                  { title: '场景大类', dataIndex: 'scene_category' },
                  { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores', render: (v: number) => formatInt(v) },
                  { title: 'GPU卡数', dataIndex: 'gpu_card_count', render: (v: number) => formatInt(v) },
                  { title: '数据盘类型', dataIndex: 'data_disk_type' },
                  { title: '数据盘数量', dataIndex: 'data_disk_count', render: (v: number) => formatInt(v) },
                  { title: '存储容量(TB)', dataIndex: 'storage_capacity_tb', render: (v: number) => formatFloat(v) },
                  { title: '功率(W)', dataIndex: 'power_watts', render: (v: number) => formatFloat(v) },
                  { title: '发布年份', dataIndex: 'release_year', render: (v: number) => formatInt(v) },
                  { title: '内存容量(GB)', dataIndex: 'memory_capacity_gb', render: (v: number) => formatFloat(v) },
                  { title: '架构标准化系数', dataIndex: 'arch_standardized_factor', render: (v: number) => formatFloat(v) }
                ]} />
              </Space>,
              '服务器管理表通过配置类型关联此表；需维护GPU卡数（GPU汇总统计依赖），以及功率/发布年份/内存容量用于后续评估。',
              <Space>
                <Button onClick={exportHostPackageTemplate}>下载导入模板</Button>
                <Upload {...makeUploadProps('packages')}><Button icon={<UploadOutlined />} loading={uploading === 'packages'}>上传并导入</Button></Upload>
              </Space>
            )
          },
          {
            key: 'cabinet',
            label: '机柜配置管理',
            children: (
              <Space direction="vertical" size="large" style={{ width: '100%' }}>
                <Card title="机柜利用率维护" extra={<Button type="primary" onClick={async () => {
                  try {
                    const resp = ensureApiOk(await updateCabinetUtilization(cabinetUtilization));
                    setCabinetUtilization(resp.data.utilization);
                    message.success('机柜利用率已保存');
                  } catch (e) {
                    message.error(parseApiError(e, '保存失败'));
                  }
                }}>保存</Button>}>
                  <Space>
                    <Text>利用率（小数）</Text>
                    <InputNumber min={0.0001} max={2} step={0.0001} value={cabinetUtilization} stringMode onChange={(v) => setCabinetUtilization(Number(v || 0))} formatter={(value) => {
                      if (value === undefined || value === null) return '';
                      const s = String(value);
                      if (!s.includes('.')) return s;
                      return s.replace(/\.?0+$/, '');
                    }} style={{ width: 180 }} />
                    <Text type="secondary">前端展示：{(cabinetUtilization * 100).toFixed(2)}%</Text>
                  </Space>
                </Card>

                <Card title="机柜配置表" extra={<Space><Button onClick={() => exportCabinetTemplate()}>下载导入模板</Button><Upload maxCount={1} accept=".xlsx" showUploadList={false} customRequest={async (options) => {
                  const file = options.file as File;
                  try {
                    const resp = ensureApiOk(await importCabinetConfigs(file));
                    message.success(`导入完成：成功 ${resp.data.success} 条`);
                    await reloadAll();
                    options.onSuccess?.({}, new XMLHttpRequest());
                  } catch (e) {
                    message.error(parseApiError(e, '导入失败'));
                    options.onError?.(new Error('import failed'));
                  }
                }}><Button icon={<UploadOutlined />}>Excel模板导入</Button></Upload><Button onClick={() => { setEditingCabinet(null); setCabinetForm({ idc: '', rated_power_kw: 0, monthly_rent: 0 }); setCabinetModalOpen(true); }}>新增机柜配置</Button></Space>}>
                  <Table rowKey="id" dataSource={cabinetRows} pagination={withTotalPagination(10)} columns={[
                    { title: '机房', dataIndex: 'idc' },
                    { title: '额定功率(KW)', dataIndex: 'rated_power_kw', render: (v: number) => formatFloat(v) },
                    { title: '机柜月租(CNY)', dataIndex: 'monthly_rent', render: (v: number) => formatFloat(v) },
                    { title: '操作', render: (_, row) => (
                      <Space>
                        <Button size="small" onClick={() => { setEditingCabinet(row); setCabinetForm({ idc: row.idc, rated_power_kw: row.rated_power_kw, monthly_rent: row.monthly_rent }); setCabinetModalOpen(true); }}>编辑</Button>
                        <Popconfirm title="确认删除该机柜配置？" onConfirm={async () => {
                          try {
                            await deleteCabinetConfig(row.id);
                            message.success('已删除');
                            await reloadAll();
                          } catch (e) {
                            message.error(parseApiError(e, '删除失败'));
                          }
                        }}>
                          <Button danger size="small">删除</Button>
                        </Popconfirm>
                      </Space>
                    )}
                  ]} />
                </Card>

                <Modal title={editingCabinet ? '编辑机柜配置' : '新增机柜配置'} open={cabinetModalOpen} onCancel={() => setCabinetModalOpen(false)} onOk={async () => {
                  try {
                    if (!cabinetForm.idc.trim()) return message.warning('请输入机房');
                    if (cabinetForm.rated_power_kw <= 0 || cabinetForm.monthly_rent <= 0) return message.warning('额定功率与机柜月租必须大于0');
                    if (editingCabinet) {
                      await updateCabinetConfig(editingCabinet.id, cabinetForm);
                    } else {
                      await createCabinetConfig(cabinetForm);
                    }
                    message.success('保存成功');
                    setCabinetModalOpen(false);
                    await reloadAll();
                  } catch (e) {
                    message.error(parseApiError(e, '保存失败'));
                  }
                }}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Input placeholder="机房" value={cabinetForm.idc} onChange={(e) => setCabinetForm({ ...cabinetForm, idc: e.target.value })} />
                    <InputNumber placeholder="额定功率(KW)" min={0.0001} value={cabinetForm.rated_power_kw} onChange={(v) => setCabinetForm({ ...cabinetForm, rated_power_kw: Number(v || 0) })} style={{ width: '100%' }} />
                    <InputNumber placeholder="机柜月租(CNY)" min={0.0001} value={cabinetForm.monthly_rent} onChange={(v) => setCabinetForm({ ...cabinetForm, monthly_rent: Number(v || 0) })} style={{ width: '100%' }} />
                  </Space>
                </Modal>
              </Space>
            )
          },
          {
            key: 'value_score_setup',
            label: '价值分管理',
            children: (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Card>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12, width: '100%' }}>
                    <Tabs
                      activeKey={valueSceneTab}
                      onChange={(k) => setValueSceneTab(k as any)}
                      items={[
                        { key: 'compute', label: `计算（${sceneScoreRows.compute.length}）` },
                        { key: 'warm_storage', label: `温存储（${sceneScoreRows.warm_storage.length}）` },
                        { key: 'hot_storage', label: `热存储（${sceneScoreRows.hot_storage.length}）` },
                        { key: 'gpu', label: `GPU（${sceneScoreRows.gpu.length}）` }
                      ]}
                      style={{ flex: '0 0 auto' }}
                    />
                    <Space style={{ marginLeft: 'auto' }}>
                      <Input.Search allowClear placeholder="配置类型" value={valueConfigKeyword} onChange={(e) => setValueConfigKeyword(e.target.value)} style={{ width: 220 }} />
                      <Button onClick={() => setValueConfigVisible((v) => !v)}>参数配置</Button>
                      <Button onClick={exportValueScoreTCO}>导出Excel</Button>
                      <Button loading={performanceLoading || tcoLoading} onClick={async () => {
                        setPerformanceLoading(true);
                        setTcoLoading(true);
                        try {
                          const [perf, tco] = await Promise.all([calculateValueScorePerformance(), calculateValueScoreTCO()]);
                          setPerformanceResult((ensureApiOk(perf) as any).data);
                          setTcoResult((ensureApiOk(tco) as any).data);
                          message.success('结果已刷新');
                        } catch (e) {
                          message.error(parseApiError(e, '刷新失败'));
                        } finally {
                          setPerformanceLoading(false);
                          setTcoLoading(false);
                        }
                      }}>刷新</Button>
                    </Space>
                  </div>
                </Card>
                {valueConfigVisible ? <>
                <Row gutter={[16, 16]}>
                  <Col xs={24} lg={12}>
                    <Card
                      title="全局参数"
                      extra={<Button onClick={reloadAll}>刷新</Button>}
                    >
                      {cabinetBaseline ? (
                        <Space direction="vertical" style={{ width: '100%' }} size="small">
                          <Text>目标机房：{cabinetBaseline.idc}</Text>
                          <Text>机柜利用率（全局参数）：{formatFloat(cabinetBaseline.cabinet_utilization)}</Text>
                          <Text>额定功率(KW)（全局参数）：{formatFloat(cabinetBaseline.min_rated_power_kw)}</Text>
                          <Text>机柜月租(CNY)（全局参数）：{formatFloat(cabinetBaseline.monthly_rent_cny)}</Text>
                          <Text>折旧月数：固定 {costParams.depreciation_months}</Text>
                          <Text>网络设备分摊成本(CNY/月)：{formatFloat(costParams.network_device_share_cny)}</Text>
                          <Text>服务器续保费(CNY/月)：{formatFloat(costParams.server_renewal_fee_cny)}</Text>
                          <Text>月TCO口径：机柜费 + 折旧 + 网络设备分摊成本 + 网络机柜等分摊 + 服务器续保费</Text>
                          <Text>机柜成本公式：{cabinetBaseline.formula}</Text>
                          <Text type="secondary">样本机柜数：{cabinetBaseline.source_count}</Text>
                          {cabinetBaseline.note ? <Text type={cabinetBaseline.status === 'warning' ? 'danger' : 'secondary'}>{cabinetBaseline.note}</Text> : null}
                        </Space>
                      ) : (
                        <Text type="secondary">暂无基线数据</Text>
                      )}
                    </Card>
                  </Col>

                  <Col xs={24} lg={12}>
                    <Card title="全局参数配置" extra={<Space><Button onClick={exportValueScorePerformanceParamsTemplate}>下载模板</Button><Upload maxCount={1} accept=".xlsx" showUploadList={false} customRequest={async (options) => {
                      const file = options.file as File;
                      try {
                        const preview = ensureApiOk(await previewValueScorePerformanceParams(file));
                        setPerformancePreview((preview as any).data);
                        ensureApiOk(await importValueScoreUnifiedParams(file));
                        const [perf, tco] = await Promise.all([calculateValueScorePerformance(), calculateValueScoreTCO()]);
                        setPerformanceResult((ensureApiOk(perf) as any).data);
                        setTcoResult((ensureApiOk(tco) as any).data);
                        message.success('配置型参数导入成功并刷新结果');
                        options.onSuccess?.({}, new XMLHttpRequest());
                      } catch (e) {
                        message.error(parseApiError(e, '导入失败'));
                        options.onError?.(new Error('import failed'));
                      }
                    }}><Button icon={<UploadOutlined />}>预检并导入Excel</Button></Upload><Button onClick={async () => {
                      try {
                        const saved = ensureApiOk(await updateValueScoreCostParams(costParams));
                        setCostParams(saved.data);
                        const tco = ensureApiOk(await calculateValueScoreTCO());
                        setTcoResult(tco.data);
                        message.success('成本参数已保存并刷新月TCO');
                      } catch (e) {
                        message.error(parseApiError(e, '保存失败'));
                      }
                    }}>保存参数</Button></Space>}>
                      <Space direction="vertical" size="small" style={{ width: '100%' }}>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>折旧月数</Text>
                          <InputNumber min={60} max={60} value={costParams.depreciation_months} disabled />
                          <Text type="secondary">固定60，不可编辑</Text>
                        </Space>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>网络设备分摊成本(CNY/月)</Text>
                          <InputNumber min={0} value={costParams.network_device_share_cny} onChange={(v) => setCostParams({ ...costParams, network_device_share_cny: Number(v || 0) })} />
                        </Space>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>服务器续保费(CNY/月)</Text>
                          <InputNumber min={0} value={costParams.server_renewal_fee_cny} onChange={(v) => setCostParams({ ...costParams, server_renewal_fee_cny: Number(v || 0) })} />
                        </Space>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>机柜利用率</Text>
                          <InputNumber min={0.0001} step={0.0001} value={costParams.cabinet_utilization} onChange={(v) => setCostParams({ ...costParams, cabinet_utilization: Number(v || 0) })} />
                        </Space>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>额定功率(KW)</Text>
                          <InputNumber min={0.0001} step={0.0001} value={costParams.rated_power_kw} onChange={(v) => setCostParams({ ...costParams, rated_power_kw: Number(v || 0) })} />
                        </Space>
                        <Space size="small" align="center">
                          <Text style={{ width: 160 }}>机柜月租(CNY)</Text>
                          <InputNumber min={0.0001} step={0.01} value={costParams.monthly_rent_cny} onChange={(v) => setCostParams({ ...costParams, monthly_rent_cny: Number(v || 0) })} />
                        </Space>
                        <Text type="secondary">当前口径：机柜费 + 折旧 + 网络设备分摊成本 + 网络机柜等分摊 + 服务器续保费</Text>
                        <Text type="secondary">单文件导入：主键配置类型。后端一次解析、一次事务落库（原值+性能参数原子提交）。</Text>
                        {performancePreview ? <Text>预检：新增 {performancePreview.new_count || 0}，更新 {performancePreview.updated_count || 0}，失败 {performancePreview.failed || 0}</Text> : null}
                        {performancePreview?.errors?.length ? <Text type="danger">失败原因：{performancePreview.errors.slice(0, 5).map((e: any) => `第${e.row}行 ${e.reason}`).join('；')}</Text> : null}
                      </Space>
                    </Card>
                  </Col>
                </Row>

                </> : null}

                <Card>
                  {performanceResult?.alert_count ? <Alert type="warning" showIcon message={`检测到 ${performanceResult.alert_count} 条告警`} description={(performanceResult.items || []).flatMap((it: any) => (it.alerts || []).map((a: any) => `${a.config_type} | ${a.error_code} | ${a.field} | ${a.current_value} | ${a.suggestion}`)).slice(0, 10).join('；') || '请检查配置类型、性能跑分、可用核数/内存等数据'} style={{ marginBottom: 12 }} /> : null}
                  <Table rowKey="config_type" dataSource={(filteredSceneScoreRows as any)[valueSceneTab] || []} scroll={{ x: 2800 }} pagination={withTotalPagination(10)} columns={[
                    { title: '配置类型', dataIndex: 'config_type', fixed: 'left', width: 160, sorter: textSorter('config_type') },
                    { title: '场景', dataIndex: 'scene_category', width: 120, sorter: textSorter('scene_category'), render: (v: string) => v || '-' },
                    { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores', sorter: numberSorter('cpu_logical_cores'), render: (v: number) => formatInt(v) },
                    { title: '内存容量(GB)', dataIndex: 'memory_capacity_gb', sorter: numberSorter('memory_capacity_gb'), render: (v: number) => formatFloat(v) },
                    { title: '存储容量(TB)', dataIndex: 'capacity_storage_tb', sorter: numberSorter('capacity_storage_tb'), render: (v: number) => formatFloat(v) },
                    { title: 'GPU卡数', dataIndex: 'count_gpu', sorter: numberSorter('count_gpu'), render: (v: number) => formatInt(v) },
                    { title: '不可用核数', dataIndex: 'unavailable_cores', sorter: numberSorter('unavailable_cores'), render: (v: number) => formatInt(v) },
                    { title: '不可用内存(GB)', dataIndex: 'unavailable_memory_gb', sorter: numberSorter('unavailable_memory_gb'), render: (v: number) => formatFloat(v) },
                    { title: '性能跑分', dataIndex: 'performance_score', sorter: numberSorter('performance_score'), render: (v: number) => formatFloat(v) },
                    { title: '可用核数', dataIndex: 'available_cores', sorter: numberSorter('available_cores'), render: (v: number) => formatInt(v) },
                    { title: '可用内存(GB)', dataIndex: 'available_memory_gb', sorter: numberSorter('available_memory_gb'), render: (v: number) => formatFloat(v) },
                    { title: '标准跑分', dataIndex: 'standard_score', sorter: numberSorter('standard_score'), render: (v: number) => formatFloat(v) },
                    { title: 'CPU性能折算系数', dataIndex: 'cpu_performance_factor', sorter: numberSorter('cpu_performance_factor'), render: (v: number) => formatFloat(v) },
                    { title: '内存配比', dataIndex: 'memory_ratio', sorter: numberSorter('memory_ratio'), render: (v: number) => formatFloat(v) },
                    { title: '内存配比系数', dataIndex: 'memory_ratio_factor', sorter: numberSorter('memory_ratio_factor'), render: (v: number) => formatFloat(v) },
                    { title: '整体性能折算比', dataIndex: 'overall_performance_ratio', sorter: numberSorter('overall_performance_ratio'), render: (v: number) => formatFloat(v) },
                    { title: '功率(W)', dataIndex: 'power_watts', sorter: numberSorter('power_watts'), render: (v: number) => formatFloat(v) },
                    { title: '功率(KW)', dataIndex: 'power_kw', sorter: numberSorter('power_kw'), render: (v: number) => formatFloat(v) },
                    { title: '机柜费/月', dataIndex: 'cabinet_cost_monthly', sorter: numberSorter('cabinet_cost_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '原值(CNY)', dataIndex: 'server_original_cny', sorter: numberSorter('server_original_cny'), render: (v: number) => formatFloat(v) },
                    { title: '折旧/月', dataIndex: 'depreciation_monthly', sorter: numberSorter('depreciation_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '网络设备分摊/月', dataIndex: 'network_device_monthly', sorter: numberSorter('network_device_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '网络机柜等分摊/月', dataIndex: 'network_cabinet_monthly', sorter: numberSorter('network_cabinet_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '服务器续保费/月', dataIndex: 'server_renewal_monthly', sorter: numberSorter('server_renewal_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '其他固定成本/月', dataIndex: 'other_fixed_cost_monthly', sorter: numberSorter('other_fixed_cost_monthly'), render: (v: number) => formatFloat(v) },
                    { title: '月TCO', dataIndex: 'total_tco_monthly', sorter: numberSorter('total_tco_monthly'), render: (v: number) => formatFloat(v) },
                    {
                      title: '单位月TCO',
                      dataIndex: 'unit_tco',
                      sorter: numberSorter('unit_tco'),
                      render: (v: number, row: any) => {
                        if (row.scene_type === 'gpu') return `${formatFloat(v)} /GPU卡`;
                        if (row.scene_type === 'warm_storage' || row.scene_type === 'hot_storage') return `${formatFloat(v)} /TB`;
                        return `${formatFloat(v)} /核`;
                      }
                    },

                    { title: '价值分v1', dataIndex: 'value_score_v1', sorter: numberSorter('value_score_v1'), render: (v: number) => formatFloat(v) }
                  ]} />
                </Card>
              </Space>
            )
          },
          {
            key: 'assets',
            label: '资产分析',
            children: (
              <Space direction="vertical" size="large" style={{ width: '100%' }}>
                <Card
                  title="国内计算服务器闲置率"
                  extra={
                    <Button icon={<DownloadOutlined />} onClick={onExportUnmatchedRackCSV} disabled={!assetAnalysis.unmatchedRackRows.length}>
                      下载未匹配清单
                    </Button>
                  }
                >
                  <Row gutter={[16, 16]}>
                    <Col xs={12} md={5}><StatisticBlock title="在用数量" value={assetAnalysis.idleSummary.active} /></Col>
                    <Col xs={12} md={5}><StatisticBlock title="闲置数量" value={assetAnalysis.idleSummary.idle} /></Col>
                    <Col xs={12} md={5}><StatisticBlock title="整备中数量" value={assetAnalysis.idleSummary.staging} /></Col>
                    <Col xs={12} md={4}><StatisticBlock title="闲置率" value={`${assetAnalysis.idleSummary.idleRate.toFixed(2)}%`} /></Col>
                    <Col xs={12} md={5}><StatisticBlock title="机柜未匹配服务器" value={assetAnalysis.idleSummary.unmatchedRack} /></Col>
                  </Row>
                  <Space direction="vertical" size={12} style={{ width: '100%', marginTop: 16 }}>
                    <Table
                      title={() => '资源统计'}
                      rowKey="category"
                      dataSource={assetAnalysis.resourceSummaryRows}
                      pagination={false}
                      size="small"
                      columns={[
                        { title: '分类', dataIndex: 'category' },
                        { title: '单位', dataIndex: 'unit', width: 90 },
                        { title: '在用算力', dataIndex: 'activeCapacity', render: (v: number, row) => formatCapacity(v, row.unit), sorter: (a, b) => a.activeCapacity - b.activeCapacity },
                        { title: '闲置算力', dataIndex: 'idleCapacity', render: (v: number, row) => formatCapacity(v, row.unit), sorter: (a, b) => a.idleCapacity - b.idleCapacity },
                        { title: '整备中', dataIndex: 'stagingCapacity', render: (v: number, row) => formatCapacity(v, row.unit), sorter: (a, b) => a.stagingCapacity - b.stagingCapacity }
                      ]}
                    />
                    <Table
                      title={() => '热存储诊断'}
                      rowKey="item"
                      dataSource={assetAnalysis.hotStorageDiagnosticRows}
                      pagination={false}
                      size="small"
                      columns={[
                        { title: '检查项', dataIndex: 'item' },
                        { title: '数量', dataIndex: 'count', width: 100, render: (v: number) => formatInt(v), sorter: (a, b) => a.count - b.count },
                        { title: '说明', dataIndex: 'note' }
                      ]}
                    />
                    <IdleStackedBarChart rows={assetAnalysis.idleRows} />
                    <Table
                      rowKey="configType"
                      dataSource={assetAnalysis.idleRows}
                      pagination={withTotalPagination(10)}
                      size="small"
                      columns={[
                        { title: '配置类型', dataIndex: 'configType' },
                        { title: '性能跑分', dataIndex: 'performanceScore', render: (v: number) => formatFloat(v), sorter: (a, b) => a.performanceScore - b.performanceScore },
                        { title: '在用数量', dataIndex: 'activeCount', render: (v: number) => formatInt(v), sorter: (a, b) => a.activeCount - b.activeCount },
                        { title: '闲置数量', dataIndex: 'idleCount', render: (v: number) => formatInt(v), sorter: (a, b) => a.idleCount - b.idleCount },
                        { title: '整备中数量', dataIndex: 'stagingCount', render: (v: number) => formatInt(v), sorter: (a, b) => a.stagingCount - b.stagingCount },
                        { title: '合计', dataIndex: 'totalCount', render: (v: number) => formatInt(v), sorter: (a, b) => a.totalCount - b.totalCount },
                        { title: '闲置率', dataIndex: 'idleRate', render: (v: number) => `${v.toFixed(2)}%`, sorter: (a, b) => a.idleRate - b.idleRate }
                      ]}
                    />
                    <Text type="secondary">口径：server.rack 关联 rack.sn 获取 rack.datacenter；机柜未匹配服务器不纳入本指标统计。国内为 rack.datacenter 非 IN 开头；闲置=PSA命中续保配置的闲置/停服PSA；整备中=未命中闲置PSA且机柜包含SPR；在用=其余。计算分类按价值分性能跑分，≥{COMPUTE_STANDARD_PERFORMANCE_THRESHOLD} 为标准计算，其它为低配计算。</Text>
                  </Space>
                </Card>
                <Card title="国内/印度保内保外概览">
                  <Table rowKey={(r) => `${r.region}-${r.snapshotDate}`} dataSource={assetAnalysis.snapshotRows} pagination={false} size="small" columns={[
                    { title: '地区', dataIndex: 'region' },
                    { title: '统计时点', dataIndex: 'snapshotLabel' },
                    { title: '日期', dataIndex: 'snapshotDate' },
                    { title: '保内', dataIndex: 'inWarranty', render: (v: number) => formatInt(v) },
                    { title: '保外', dataIndex: 'outWarranty', render: (v: number) => formatInt(v) },
                    { title: '总数量', dataIndex: 'total', render: (v: number) => formatInt(v) },
                    { title: '累计过保占比', dataIndex: 'outWarrantyRatio', render: (v: number) => `${v.toFixed(2)}%` }
                  ]} />
                  <Text type="secondary">口径：IDC 以 "IN" 开头判定为印度，其余归为国内；日期为空/异常按已过保处理。</Text>
                </Card>
                <Card title="国内服务器过保趋势图"><AssetTrendChart points={assetAnalysis.trends.domestic} total={assetAnalysis.totals.domestic} regionLabel="国内" /></Card>
                <Card title="印度服务器过保趋势图"><AssetTrendChart points={assetAnalysis.trends.india} total={assetAnalysis.totals.india} regionLabel="印度" /></Card>
              </Space>
            )
          }
        ].filter((item) => visibleTabKeys.includes(item.key as DataKey))}
      />
    </Space>
  );
}

type RegionKey = 'domestic' | 'india';
interface AssetSnapshotRow { region: '国内' | '印度'; snapshotLabel: string; snapshotDate: string; inWarranty: number; outWarranty: number; total: number; outWarrantyRatio: number; }
interface AssetTrendPoint { year: number; outCount: number; cumulativeOutCount: number; cumulativeOutRatio: number; }
interface AssetServerRow { sn: string; psa: string; configType: string; rack: string; idc: string; warrantyEndDate: string; }
interface AssetIdleRow { configType: string; activeCount: number; idleCount: number; stagingCount: number; totalCount: number; idleRate: number; performanceScore: number; }
interface AssetIdleSummary { active: number; idle: number; staging: number; idleRate: number; unmatchedRack: number; }
interface AssetUnmatchedRackRow { sn: string; psa: string; configType: string; rack: string; idc: string; reason: string; }
interface AssetResourceRow { category: string; unit: string; activeCapacity: number; idleCapacity: number; stagingCapacity: number; }
interface AssetResourceDiagnosticRow { item: string; count: number; note: string; }
interface AssetAnalysisInput {
  legacyServers: ServerItem[];
  metaServers: MetaRecord[];
  metaRacks: MetaRecord[];
  metaConfigs: MetaRecord[];
  valuePackages: HostPackageConfig[];
  performanceItems: ValueScorePerformanceCalcItem[];
  idleStoppedPSAs: string[];
}

function buildAssetAnalysis(input: AssetAnalysisInput) {
  const now = new Date();
  const nextYear0630 = new Date(now.getFullYear() + 1, 5, 30);
  const snapshots = [{ label: '当前时间', date: now }, { label: '次年6月30日', date: nextYear0630 }];
  const totals: Record<RegionKey, number> = { domestic: 0, india: 0 };
  const snapshotRows: AssetSnapshotRow[] = [];
  const trends: Record<RegionKey, AssetTrendPoint[]> = { domestic: [], india: [] };
  const assetServers = normalizeAssetServers(input);

  (['domestic', 'india'] as RegionKey[]).forEach((region) => {
    const list = assetServers.filter((s) => resolveRegion(s.idc) === region);
    totals[region] = list.length;
    for (const snap of snapshots) {
      let outWarranty = 0;
      for (const item of list) {
        const end = parseYMD(item.warrantyEndDate);
        if (!end || end.getTime() < snap.date.getTime()) outWarranty += 1;
      }
      const total = list.length;
      snapshotRows.push({ region: region === 'domestic' ? '国内' : '印度', snapshotLabel: snap.label, snapshotDate: formatDate(snap.date), inWarranty: Math.max(0, total - outWarranty), outWarranty, total, outWarrantyRatio: total > 0 ? (outWarranty * 100) / total : 0 });
    }
    const yearMap = new Map<number, number>();
    for (const item of list) {
      const end = parseYMD(item.warrantyEndDate);
      if (!end) continue;
      const y = end.getFullYear();
      yearMap.set(y, (yearMap.get(y) || 0) + 1);
    }
    let cumulative = 0;
    trends[region] = [...yearMap.entries()].sort((a, b) => a[0] - b[0]).map(([year, outCount]) => {
      cumulative += outCount;
      return { year, outCount, cumulativeOutCount: cumulative, cumulativeOutRatio: list.length > 0 ? (cumulative * 100) / list.length : 0 };
    });
  });

  const idleAnalysis = buildIdleAnalysis(assetServers, input);
  const resourceSummaryRows = buildResourceSummary(assetServers, input);
  const hotStorageDiagnosticRows = buildHotStorageDiagnostic(assetServers, input);
  return { snapshotRows, trends, totals, resourceSummaryRows, hotStorageDiagnosticRows, ...idleAnalysis };
}

function normalizeAssetServers(input: AssetAnalysisInput): AssetServerRow[] {
  if (input.metaServers.length) {
    const rackIDC = new Map<string, string>();
    for (const r of input.metaRacks) {
      const data = r.data || {};
      const rackNo = pickMeta(data, 'sn', 'rack', '机柜编号');
      if (rackNo) rackIDC.set(rackNo, pickMeta(data, 'datacenter', 'idc', '机房'));
    }
    return input.metaServers.map((r) => {
      const data = r.data || {};
      const rack = pickMeta(data, 'rack', '机柜', '机柜编号');
      return {
        sn: pickMeta(data, 'sn', 'SN'),
        psa: pickMeta(data, 'psa', 'PSA'),
        configType: pickMeta(data, 'config_type', '配置类型'),
        rack,
        idc: rackIDC.get(rack) || '',
        warrantyEndDate: pickMeta(data, 'server_warranty_last_date', 'warranty_end_date', '保修结束日期')
      };
    });
  }
  return input.legacyServers.map((s) => ({
    sn: s.sn,
    psa: s.psa,
    configType: s.config_type,
    rack: '',
    idc: s.idc || '',
    warrantyEndDate: s.warranty_end_date || ''
  }));
}

function buildIdleAnalysis(assetServers: AssetServerRow[], input: AssetAnalysisInput): { idleRows: AssetIdleRow[]; idleSummary: AssetIdleSummary; unmatchedRackRows: AssetUnmatchedRackRow[] } {
  const computeInfoByConfig = buildComputeClassification(input);
  const idlePatterns = normalizePSAPatterns(input.idleStoppedPSAs);
  const byConfig = new Map<string, AssetIdleRow>();
  const unmatchedRackRows: AssetUnmatchedRackRow[] = [];
  let unmatchedRack = 0;
  for (const s of assetServers) {
    const configType = s.configType.trim();
    const computeInfo = computeInfoByConfig.get(configType);
    if (!configType || !computeInfo) continue;
    if (!s.rack || !s.idc) {
      unmatchedRack += 1;
      unmatchedRackRows.push({
        sn: s.sn,
        psa: s.psa,
        configType,
        rack: s.rack,
        idc: s.idc,
        reason: !s.rack ? '服务器机柜为空' : '机柜未匹配到机房'
      });
      continue;
    }
    if (resolveRegion(s.idc) !== 'domestic') continue;

    const row = byConfig.get(configType) || {
      configType,
      activeCount: 0,
      idleCount: 0,
      stagingCount: 0,
      totalCount: 0,
      idleRate: 0,
      performanceScore: computeInfo.performanceScore
    };
    const isIdle = matchPSA(s.psa, idlePatterns);
    const isStaging = !isIdle && s.rack.toUpperCase().includes('SPR');
    const isActive = !isIdle && !isStaging;
    if (isIdle) row.idleCount += 1;
    if (isStaging) row.stagingCount += 1;
    if (isActive) row.activeCount += 1;
    row.totalCount = row.activeCount + row.idleCount + row.stagingCount;
    row.idleRate = row.totalCount > 0 ? (row.idleCount * 100) / row.totalCount : 0;
    byConfig.set(configType, row);
  }

  const idleRows = [...byConfig.values()].sort((a, b) => {
    if (b.performanceScore !== a.performanceScore) return b.performanceScore - a.performanceScore;
    return a.configType.localeCompare(b.configType, 'zh-Hans-CN');
  });
  const active = idleRows.reduce((sum, r) => sum + r.activeCount, 0);
  const idle = idleRows.reduce((sum, r) => sum + r.idleCount, 0);
  const staging = idleRows.reduce((sum, r) => sum + r.stagingCount, 0);
  return {
    idleRows,
    unmatchedRackRows,
    idleSummary: {
      active,
      idle,
      staging,
      unmatchedRack,
      idleRate: active + idle + staging > 0 ? (idle * 100) / (active + idle + staging) : 0
    }
  };
}

function buildComputeClassification(input: AssetAnalysisInput) {
  const performanceByConfig = new Map<string, number>();
  for (const item of input.performanceItems) {
    performanceByConfig.set(String(item.config_type || '').trim(), Number(item.performance_score || 0));
  }

  const computeInfoByConfig = new Map<string, { performanceScore: number; className: string }>();
  const addIfCompute = (configType: string, scene: string) => {
    const key = String(configType || '').trim();
    if (!key || !isComputeScene(scene)) return;
    const performanceScore = performanceByConfig.get(key) || 0;
    computeInfoByConfig.set(key, {
      performanceScore,
      className: performanceScore >= COMPUTE_STANDARD_PERFORMANCE_THRESHOLD ? '标准计算' : '低配计算'
    });
  };

  for (const pkg of input.valuePackages) {
    addIfCompute(pkg.config_type, String((pkg as any).application_category || pkg.scene_category || '').trim());
  }
  for (const r of input.metaConfigs) {
    const data = r.data || {};
    addIfCompute(pickMeta(data, 'config_type', '配置类型'), pickMeta(data, 'application_category', 'scene_category', '场景大类'));
  }
  if (!computeInfoByConfig.size) {
    for (const item of input.performanceItems) {
      const key = String(item.config_type || '').trim();
      if (!key) continue;
      const performanceScore = Number(item.performance_score || 0);
      computeInfoByConfig.set(key, {
        performanceScore,
        className: performanceScore >= COMPUTE_STANDARD_PERFORMANCE_THRESHOLD ? '标准计算' : '低配计算'
      });
    }
  }
  return computeInfoByConfig;
}

function buildResourceSummary(assetServers: AssetServerRow[], input: AssetAnalysisInput): AssetResourceRow[] {
  const computeInfoByConfig = buildComputeClassification(input);
  const packageMap = new Map<string, HostPackageConfig>();
  for (const pkg of input.valuePackages) packageMap.set(String(pkg.config_type || '').trim(), pkg);

  const idlePatterns = normalizePSAPatterns(input.idleStoppedPSAs);
  const summaryMap = new Map<string, AssetResourceRow>();

  const getRow = (category: string, unit: string) => {
    const existing = summaryMap.get(category);
    if (existing) return existing;
    const created = { category, unit, activeCapacity: 0, idleCapacity: 0, stagingCapacity: 0 };
    summaryMap.set(category, created);
    return created;
  };

  for (const s of assetServers) {
    if (resolveRegion(s.idc) !== 'domestic') continue;
    if (!s.rack || !s.idc) continue;
    const configType = s.configType.trim();
    if (!configType) continue;
    const pkg = packageMap.get(configType);
    if (!pkg) continue;

    const isIdle = matchPSA(s.psa, idlePatterns);
    const isStaging = !isIdle && s.rack.toUpperCase().includes('SPR');
    const isActive = !isIdle && !isStaging;

    if (computeInfoByConfig.has(configType)) {
      const row = getRow(computeInfoByConfig.get(configType)?.className || '低配计算', '核');
      const capacity = Number(pkg.cpu_logical_cores || 0);
      if (isIdle) row.idleCapacity += capacity;
      else if (isStaging) row.stagingCapacity += capacity;
      else if (isActive) row.activeCapacity += capacity;
      continue;
    }

    if (pkg.scene_category && isStorageScene(pkg.scene_category)) {
      const row = getRow(normalizeStorageCategory(pkg.scene_category), 'TB');
      const capacity = Number(pkg.storage_capacity_tb || 0);
      if (isIdle) row.idleCapacity += capacity;
      else if (isStaging) row.stagingCapacity += capacity;
      else if (isActive) row.activeCapacity += capacity;
      continue;
    }

    if (Number(pkg.gpu_card_count || 0) > 0 || isGPUScene(pkg.scene_category)) {
      const row = getRow('GPU', '卡');
      const capacity = Number(pkg.gpu_card_count || 0);
      if (isIdle) row.idleCapacity += capacity;
      else if (isStaging) row.stagingCapacity += capacity;
      else if (isActive) row.activeCapacity += capacity;
    }
  }

  const order = [
    { category: '低配计算', unit: '核' },
    { category: '标准计算', unit: '核' },
    { category: '温存储', unit: 'TB' },
    { category: '热存储', unit: 'TB' },
    { category: 'GPU', unit: '卡' }
  ];
  return order.map(({ category, unit }) => summaryMap.get(category) || { category, unit, activeCapacity: 0, idleCapacity: 0, stagingCapacity: 0 });
}

function buildHotStorageDiagnostic(assetServers: AssetServerRow[], input: AssetAnalysisInput): AssetResourceDiagnosticRow[] {
  const packageMap = new Map<string, HostPackageConfig>();
  for (const pkg of input.valuePackages) packageMap.set(String(pkg.config_type || '').trim(), pkg);

  const knownHotConfigs = new Set<string>();
  for (const pkg of input.valuePackages) {
    if (isHotStorageScene(pkg.scene_category)) knownHotConfigs.add(String(pkg.config_type || '').trim());
  }
  for (const r of input.metaConfigs) {
    const data = r.data || {};
    const configType = pickMeta(data, 'config_type', '配置类型');
    const scene = pickMeta(data, 'application_category', 'scene_category', '场景大类');
    if (isHotStorageScene(scene)) knownHotConfigs.add(configType.trim());
  }
  knownHotConfigs.delete('');

  let included = 0;
  let noPackage = 0;
  let notHotScene = 0;
  let unmatchedRackOrIDC = 0;
  let nonDomestic = 0;
  let zeroCapacity = 0;
  for (const s of assetServers) {
    const configType = s.configType.trim();
    if (!configType) continue;
    const pkg = packageMap.get(configType);
    const knownAsHot = knownHotConfigs.has(configType);
    if (!pkg) {
      if (knownAsHot) noPackage += 1;
      continue;
    }
    if (!isHotStorageScene(pkg.scene_category)) {
      if (knownAsHot || maybeHotStorageScene(pkg.scene_category)) notHotScene += 1;
      continue;
    }
    if (!s.rack || !s.idc) {
      unmatchedRackOrIDC += 1;
      continue;
    }
    if (resolveRegion(s.idc) !== 'domestic') {
      nonDomestic += 1;
      continue;
    }
    if (Number(pkg.storage_capacity_tb || 0) <= 0) {
      zeroCapacity += 1;
      continue;
    }
    included += 1;
  }

  return [
    { item: '热存储套餐配置', count: knownHotConfigs.size, note: '主机套餐/配置元数据中场景为热存储的配置类型数量' },
    { item: '已纳入热存储服务器', count: included, note: '同时满足配置匹配、热存储场景、国内、机柜/机房匹配、容量>0' },
    { item: '配置未匹配套餐', count: noPackage, note: '服务器配置类型没有在主机套餐中找到同名配置' },
    { item: '套餐场景未识别为热存储', count: notHotScene, note: '配置看起来属于热存储，但套餐场景值没有被识别为热存储' },
    { item: '机柜或机房未匹配', count: unmatchedRackOrIDC, note: '服务器 rack 为空，或 rack 未关联到机房 idc' },
    { item: '非国内服务器', count: nonDomestic, note: '机房 idc 以 IN 开头，按印度排除' },
    { item: '容量为0', count: zeroCapacity, note: '热存储套餐存储容量(TB)为空或小于等于0' }
  ];
}

function pickMeta(data: Record<string, any>, ...keys: string[]) {
  for (const key of keys) {
    const v = data[key];
    const text = metaValueToText(v);
    if (text) return text;
  }
  return '';
}

function metaValueToText(value: any): string {
  if (value === undefined || value === null) return '';
  if (Array.isArray(value)) {
    for (const item of value) {
      const text = metaValueToText(item);
      if (text) return text;
    }
    return '';
  }
  if (typeof value === 'object') {
    for (const key of ['sn', 'SN', 'rack', 'config_type', 'value', 'label', 'name', 'text', 'id']) {
      const text = metaValueToText(value[key]);
      if (text) return text;
    }
    return '';
  }

  const raw = String(value).trim();
  if (!raw) return '';
  if ((raw.startsWith('{') && raw.endsWith('}')) || (raw.startsWith('[') && raw.endsWith(']'))) {
    try {
      const parsed = JSON.parse(raw);
      const text = metaValueToText(parsed);
      if (text) return text;
    } catch {
      // Use raw text when it is not valid JSON.
    }
  }
  return raw;
}

function isComputeScene(raw: string) {
  const scene = String(raw || '').trim().toLowerCase();
  return ['计算', '计算型', 'compute'].includes(scene);
}

function isStorageScene(raw: string) {
  return isWarmStorageScene(raw) || isHotStorageScene(raw);
}

function isGPUScene(raw?: string) {
  const scene = normalizeSceneToken(raw);
  return ['gpu'].includes(scene);
}

function normalizeStorageCategory(raw: string) {
  if (isHotStorageScene(raw)) return '热存储';
  return '温存储';
}

function isWarmStorageScene(raw?: string) {
  const scene = normalizeSceneToken(raw);
  return ['温存储', '温储', '温', 'warmstorage', 'coldstorage', 'wenstorage'].includes(scene);
}

function isHotStorageScene(raw?: string) {
  const scene = normalizeSceneToken(raw);
  return ['热存储', '热储', '热', 'hotstorage'].includes(scene);
}

function maybeHotStorageScene(raw?: string) {
  const scene = normalizeSceneToken(raw);
  return scene.includes('hot') || scene.includes('热');
}

function normalizeSceneToken(raw?: string) {
  return String(raw || '').trim().toLowerCase().replace(/[\s_-]/g, '');
}

function normalizePSAPatterns(values: string[]) {
  return values.flatMap(splitPSATokens);
}

function normalizePSA(v: string) {
  let n = String(v || '').trim().toLowerCase();
  n = n.replace(/^["']+|["']+$/g, '');
  n = n.replace(/\\/g, '/');
  n = n.replace(/\s+/g, '');
  n = n.replace(/\/+$/, '');
  return n;
}

function splitPSATokens(raw: string) {
  const text = String(raw || '').trim();
  if (!text) return [];

  if (text.startsWith('[') && text.endsWith(']')) {
    try {
      const arr = JSON.parse(text);
      if (Array.isArray(arr)) {
        const out = arr.map((x) => normalizePSA(String(x || ''))).filter(Boolean);
        if (out.length) return out;
      }
    } catch {
      // Fall through to delimiter based parsing.
    }
  }

  return text.split(/[,\n，;；\t]/).map(normalizePSA).filter(Boolean);
}

function matchPSA(raw: string, patterns: string[]) {
  const tokens = splitPSATokens(raw);
  if (!tokens.length || !patterns.length) return false;
  return tokens.some((token) => patterns.some((pattern) => token === pattern || token.startsWith(`${pattern}/`)));
}

function StatisticBlock({ title, value }: { title: string; value: string | number }) {
  return (
    <div style={{ border: '1px solid #f0f0f0', borderRadius: 6, padding: 12, background: '#fff' }}>
      <Text type="secondary">{title}</Text>
      <div style={{ marginTop: 6, fontSize: 22, fontWeight: 600 }}>{typeof value === 'number' ? formatInt(value) : value}</div>
    </div>
  );
}

function formatCapacity(v: number, unit: string) {
  if (unit === 'TB') return formatFloat(v);
  return formatInt(v);
}

function IdleStackedBarChart({ rows }: { rows: AssetIdleRow[] }) {
  if (!rows.length) return <Text type="secondary">暂无可用于绘图的国内计算服务器闲置数据。</Text>;
  const width = 920;
  const height = 340;
  const m = { left: 52, right: 24, top: 28, bottom: 78 };
  const innerW = width - m.left - m.right;
  const innerH = height - m.top - m.bottom;
  const maxTotal = Math.max(1, ...rows.map((r) => r.totalCount));
  const barW = Math.max(18, Math.min(42, innerW / Math.max(rows.length, 1) * 0.52));
  const x = (idx: number) => m.left + ((idx + 0.5) * innerW) / rows.length;
  const barH = (v: number) => (v / maxTotal) * innerH;
  return (
    <svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', background: '#fff', border: '1px solid #f0f0f0', borderRadius: 6 }}>
      <g>
        <rect x={m.left} y={8} width="12" height="12" fill="#69b1ff" rx="2" />
        <text x={m.left + 18} y={18} fontSize="11" fill="#666">在用</text>
        <rect x={m.left + 72} y={8} width="12" height="12" fill="#ffa940" rx="2" />
        <text x={m.left + 90} y={18} fontSize="11" fill="#666">闲置</text>
        <rect x={m.left + 144} y={8} width="12" height="12" fill="#95de64" rx="2" />
        <text x={m.left + 162} y={18} fontSize="11" fill="#666">整备中</text>
      </g>
      {[0, 0.25, 0.5, 0.75, 1].map((p) => {
        const y = m.top + innerH - p * innerH;
        return <g key={p}><line x1={m.left} x2={width - m.right} y1={y} y2={y} stroke="#f0f0f0" /><text x={10} y={y + 4} fontSize="10" fill="#888">{formatInt(maxTotal * p)}</text></g>;
      })}
      {rows.map((row, i) => {
        const activeH = barH(row.activeCount);
        const idleH = barH(row.idleCount);
        const stagingH = barH(row.stagingCount);
        const baseY = m.top + innerH;
        const activeY = baseY - activeH;
        const idleY = activeY - idleH;
        const stagingY = idleY - stagingH;
        return (
          <g key={row.configType}>
            <rect x={x(i) - barW / 2} y={activeY} width={barW} height={activeH} fill="#69b1ff" rx="3" />
            <rect x={x(i) - barW / 2} y={idleY} width={barW} height={idleH} fill="#ffa940" rx="3" />
            <rect x={x(i) - barW / 2} y={stagingY} width={barW} height={stagingH} fill="#95de64" rx="3" />
            <text x={x(i)} y={Math.max(14, stagingY - 6)} textAnchor="middle" fontSize="10" fill="#595959">{formatInt(row.totalCount)}</text>
            <text x={x(i)} y={height - 42} textAnchor="end" fontSize="10" fill="#666" transform={`rotate(-35 ${x(i)} ${height - 42})`}>{row.configType}</text>
            <title>{`${row.configType}：在用 ${row.activeCount}，闲置 ${row.idleCount}，整备中 ${row.stagingCount}，闲置率 ${row.idleRate.toFixed(2)}%，性能跑分 ${formatFloat(row.performanceScore)}`}</title>
          </g>
        );
      })}
    </svg>
  );
}
function resolveRegion(idc?: string): RegionKey { const norm = (idc || '').trim().toUpperCase(); return norm.startsWith('IN') ? 'india' : 'domestic'; }
function parseYMD(v?: string) { if (!v) return null; const m = /^\s*(\d{4})-(\d{2})-(\d{2})\s*$/.exec(v); if (!m) return null; const y = Number(m[1]); const mon = Number(m[2]); const d = Number(m[3]); if (!Number.isFinite(y) || mon < 1 || mon > 12 || d < 1 || d > 31) return null; return new Date(y, mon - 1, d); }
function formatDate(d: Date) { const y = d.getFullYear(); const m = `${d.getMonth() + 1}`.padStart(2, '0'); const day = `${d.getDate()}`.padStart(2, '0'); return `${y}-${m}-${day}`; }
function AssetTrendChart({ points, total, regionLabel }: { points: AssetTrendPoint[]; total: number; regionLabel: string }) {
  if (!points.length) return <Text type="secondary">暂无可用于绘图的过保日期数据。</Text>;
  const width = 880; const height = 320; const m = { left: 52, right: 54, top: 28, bottom: 54 }; const innerW = width - m.left - m.right; const innerH = height - m.top - m.bottom; const maxCount = Math.max(1, ...points.map((p) => p.outCount));
  const x = (idx: number) => m.left + ((idx + 0.5) * innerW) / points.length; const barW = Math.max(14, Math.min(42, innerW / Math.max(1, points.length) * 0.55)); const yCount = (v: number) => m.top + innerH - (v / maxCount) * innerH; const yRatio = (v: number) => m.top + innerH - (v / 100) * innerH;
  const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i)},${yRatio(p.cumulativeOutRatio)}`).join(' ');
  const labelMinGap = 10; const labelTop = m.top + 2; const labelBottom = m.top + innerH - 2; const clampLabelY = (v: number) => Math.max(labelTop, Math.min(labelBottom, v));
  const labelPositions = points.map((p) => { const nodeY = yRatio(p.cumulativeOutRatio); let barY = clampLabelY(yCount(p.outCount) - 6); let lineY = clampLabelY(nodeY - labelMinGap); if (lineY > nodeY - labelMinGap) lineY = nodeY - labelMinGap; lineY = clampLabelY(lineY); if (Math.abs(barY - lineY) < labelMinGap) { if (barY <= lineY) barY = clampLabelY(lineY - labelMinGap); else lineY = clampLabelY(barY - labelMinGap); } return { barY, lineY }; });
  return (<Space direction="vertical" size={12} style={{ width: '100%' }}><svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', background: '#fff', borderRadius: 8 }}><g><rect x={m.left} y={6} width="12" height="12" fill="#91caff" rx="2" /><text x={m.left + 18} y={16} fontSize="11" fill="#666">当年过保数量（柱状）</text><line x1={m.left + 170} y1={12} x2={m.left + 194} y2={12} stroke="#ff4d4f" strokeWidth="2.5" /><circle cx={m.left + 182} cy={12} r="3.5" fill="#ff4d4f" /><text x={m.left + 202} y={16} fontSize="11" fill="#666">累计过保占比（曲线）</text></g>{[0, 25, 50, 75, 100].map((r) => (<g key={r}><line x1={m.left} x2={width - m.right} y1={yRatio(r)} y2={yRatio(r)} stroke="#f0f0f0" strokeWidth="1" /><text x={width - m.right + 6} y={yRatio(r) + 4} fontSize="10" fill="#888">{r}%</text></g>))}{points.map((p, i) => (<g key={p.year}><rect x={x(i) - barW / 2} y={yCount(p.outCount)} width={barW} height={Math.max(0, m.top + innerH - yCount(p.outCount))} fill="#91caff" rx="4"><title>{`${p.year}年过保数量：${p.outCount}`}</title></rect><text x={x(i)} y={labelPositions[i].barY} textAnchor="middle" fontSize="10" fill="#3b6ea8">{formatInt(p.outCount)}</text><text x={x(i)} y={height - 22} textAnchor="middle" fontSize="11" fill="#666">{p.year}</text></g>))}<path d={linePath} fill="none" stroke="#ff4d4f" strokeWidth="2.5" />{points.map((p, i) => (<g key={`dot-${p.year}`}><circle cx={x(i)} cy={yRatio(p.cumulativeOutRatio)} r="4" fill="#ff4d4f" /><text x={x(i)} y={labelPositions[i].lineY} textAnchor="middle" fontSize="10" fill="#c62828">{`${p.cumulativeOutRatio.toFixed(2)}%`}</text><title>{`${p.year}年累计过保占比：${p.cumulativeOutRatio.toFixed(2)}%`}</title></g>))}<text x={m.left - 40} y={m.top - 2} fontSize="10" fill="#888">过保数量</text><text x={width - m.right + 6} y={m.top - 2} fontSize="10" fill="#888">累计占比</text></svg><Table size="small" pagination={false} rowKey={(r) => `${regionLabel}-${r.year}`} dataSource={points} columns={[{ title: '年份', dataIndex: 'year' }, { title: '当年过保数量', dataIndex: 'outCount', render: (v: number) => formatInt(v) }, { title: '累计过保数量', dataIndex: 'cumulativeOutCount', render: (v: number) => formatInt(v) }, { title: '累计过保占比', dataIndex: 'cumulativeOutRatio', render: (v: number) => `${v.toFixed(2)}%` }]} /><Text type="secondary">{regionLabel}样本总量：{formatInt(total)} 台</Text></Space>);
}
function withTotalPagination(pageSize: number) { return { pageSize, showTotal: (total: number) => `共${total}条，${Math.ceil(total / pageSize)}页` }; }
function csvCell(v: string | number | undefined | null) { return `"${String(v ?? '').replace(/"/g, '""')}"`; }
function downloadCSV(filename: string, csv: string) {
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
function formatDateTime(d: Date) {
  return `${formatDate(d)}-${`${d.getHours()}`.padStart(2, '0')}${`${d.getMinutes()}`.padStart(2, '0')}${`${d.getSeconds()}`.padStart(2, '0')}`;
}
function formatInt(v?: number) { return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 }); }
function formatFloat(v?: number) { return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }); }
function formatMaybeNumber(v?: string) { const n = Number((v || '').trim()); if (Number.isNaN(n)) return v || '-'; return n.toLocaleString('en-US', { maximumFractionDigits: 2 }); }
