import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { Alert, Button, Card, Col, Input, InputNumber, Modal, Popconfirm, Row, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import type { UploadProps } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import {
  createCabinetConfig,
  deleteCabinetConfig,
  exportServerPackageAnomalies,
  exportHostPackageTemplate,
  exportValueScoreCostParamsTemplate,
  exportValueScoreOriginalValueTemplate,
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
  importValueScoreCostParams,
  importValueScoreOriginalValues,
  importValueScorePerformanceParams,
  previewValueScorePerformanceParams,
  listValueScorePerformanceParams,
  exportCabinetTemplate,
  listCabinetConfigs,
  listHostPackages,
  listServers,
  updateCabinetConfig,
  updateCabinetUtilization
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  CabinetConfig,
  HostPackageConfig,
  ImportResult,
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
  value_score_setup: '价值评分配置底座',
  assets: '资产分析'
};

export default function ImportPage() {
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
  const [costParams, setCostParams] = useState<ValueScoreCostParams>({ depreciation_months: 60, network_device_share_cny: 0, server_renewal_fee_cny: 0 });
  const [performancePreview, setPerformancePreview] = useState<any>(null);
  const [performanceResult, setPerformanceResult] = useState<{ items: ValueScorePerformanceCalcItem[]; alert_count: number; note?: string } | null>(null);
  const [tcoResult, setTcoResult] = useState<ValueScoreTCOResult | null>(null);
  const [tcoLoading, setTcoLoading] = useState(false);
  const [performanceLoading, setPerformanceLoading] = useState(false);

  async function reloadAll() {
    try {
      const [s1, s2, s3, s4, s5, s6, s7, s8] = await Promise.all([
        listServers(),
        listHostPackages(),
        getCabinetUtilization(),
        listCabinetConfigs(),
        getValueScoreCabinetBaseline(),
        getValueScoreCostParams(),
        listValueScorePerformanceParams(),
        calculateValueScorePerformance()
      ]);
      setServers((ensureApiOk(s1) as any).data.list);
      setPackages((ensureApiOk(s2) as any).data.list);
      setCabinetUtilization((ensureApiOk(s3) as any).data.utilization || 1);
      setCabinetRows((ensureApiOk(s4) as any).data.list || []);
      setCabinetBaseline((ensureApiOk(s5) as any).data);
      setCostParams((ensureApiOk(s6) as any).data);
      setPerformancePreview((ensureApiOk(s7) as any).data);
      setPerformanceResult((ensureApiOk(s8) as any).data);
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
      x.server_value_score,
      x.arch_standardized_factor
    ].some((v) => String(v ?? '').toLowerCase().includes(q)));
  }, [packages, packageKeyword]);

  const assetAnalysis = useMemo(() => buildAssetAnalysis(servers), [servers]);

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
                  { title: '服务器价值分', dataIndex: 'server_value_score', render: (v: number) => formatFloat(v) },
                  { title: '架构标准化系数', dataIndex: 'arch_standardized_factor', render: (v: number) => formatFloat(v) }
                ]} />
              </Space>,
              '服务器管理表通过配置类型关联此表；需维护服务器价值分（PSA非数字时基准）、GPU卡数（GPU汇总统计依赖），以及功率/发布年份/内存容量用于后续评估。',
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
            label: '价值评分配置底座',
            children: (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Row gutter={[16, 16]}>
                  <Col xs={24} lg={12}>
                    <Card
                      title="价值评分参数基线（只读）"
                      extra={<Button onClick={reloadAll}>刷新</Button>}
                    >
                      {cabinetBaseline ? (
                        <Space direction="vertical" style={{ width: '100%' }} size="small">
                          <Text>目标机房：{cabinetBaseline.idc}</Text>
                          <Text>机柜利用率（来自机柜配置管理）：{formatFloat(cabinetBaseline.cabinet_utilization)}</Text>
                          <Text>最低额定功率(KW)：{formatFloat(cabinetBaseline.min_rated_power_kw)}</Text>
                          <Text>对应机柜月租(CNY)：{formatFloat(cabinetBaseline.monthly_rent_cny)}</Text>
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
                    <Card title="3.1.1 成本参数配置" extra={<Space><Button onClick={exportValueScoreCostParamsTemplate}>下载导入模板</Button><Upload maxCount={1} accept=".xlsx" showUploadList={false} customRequest={async (options) => {
                      const file = options.file as File;
                      try {
                        const resp = ensureApiOk(await importValueScoreCostParams(file));
                        setCostParams(resp.data);
                        const tco = ensureApiOk(await calculateValueScoreTCO());
                        setTcoResult(tco.data);
                        message.success('成本参数导入成功并刷新月TCO');
                        options.onSuccess?.({}, new XMLHttpRequest());
                      } catch (e) {
                        message.error(parseApiError(e, '导入失败'));
                        options.onError?.(new Error('import failed'));
                      }
                    }}><Button icon={<UploadOutlined />}>导入Excel</Button></Upload><Button onClick={async () => {
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
                        <Text type="secondary">当前口径：机柜费 + 折旧 + 网络设备分摊成本 + 网络机柜等分摊 + 服务器续保费</Text>
                      </Space>
                    </Card>
                  </Col>
                </Row>

                <Card title="3.1.2 原值导入（按配置类型）" extra={<Space><Button onClick={exportValueScoreOriginalValueTemplate}>下载导入模板</Button><Upload maxCount={1} accept=".xlsx" showUploadList={false} customRequest={async (options) => {
                  const file = options.file as File;
                  try {
                    ensureApiOk(await importValueScoreOriginalValues(file));
                    const tco = ensureApiOk(await calculateValueScoreTCO());
                    setTcoResult(tco.data);
                    message.success('原值导入成功并刷新月TCO');
                    options.onSuccess?.({}, new XMLHttpRequest());
                  } catch (e) {
                    message.error(parseApiError(e, '导入失败'));
                    options.onError?.(new Error('import failed'));
                  }
                }}><Button icon={<UploadOutlined />}>导入Excel</Button></Upload></Space>}>
                  <Text type="secondary">模板仅包含“配置类型、原值(CNY)”两列；用于按套餐维护原值。</Text>
                </Card>

                <Card title="3.2 服务器性能参数配置" extra={<Space><Button onClick={exportValueScorePerformanceParamsTemplate}>下载导入模板</Button><Upload maxCount={1} accept=".xlsx" showUploadList={false} customRequest={async (options) => {
                  const file = options.file as File;
                  try {
                    const preview = ensureApiOk(await previewValueScorePerformanceParams(file));
                    setPerformancePreview((preview as any).data);
                    ensureApiOk(await importValueScorePerformanceParams(file));
                    const perf = ensureApiOk(await calculateValueScorePerformance());
                    setPerformanceResult((perf as any).data);
                    message.success('性能参数导入成功');
                    options.onSuccess?.({}, new XMLHttpRequest());
                  } catch (e) {
                    message.error(parseApiError(e, '导入失败'));
                    options.onError?.(new Error('import failed'));
                  }
                }}><Button icon={<UploadOutlined />}>预检并导入Excel</Button></Upload><Button onClick={async () => {
                  try {
                    const perf = ensureApiOk(await calculateValueScorePerformance());
                    setPerformanceResult((perf as any).data);
                    message.success('性能折算已刷新');
                  } catch (e) {
                    message.error(parseApiError(e, '刷新失败'));
                  }
                }}>刷新折算</Button></Space>}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Text type="secondary">导入主键：配置类型；字段：不可用核数、不可用内存容量(GB)、性能跑分。</Text>
                    {performancePreview ? <Text>预检：新增 {performancePreview.new_count || 0}，更新 {performancePreview.updated_count || 0}，失败 {performancePreview.failed || 0}</Text> : null}
                    {performanceResult ? <Text>告警数：{performanceResult.alert_count}</Text> : null}
                  </Space>
                </Card>

                <Card title="3.2 性能折算结果" extra={<Space><Button onClick={exportValueScorePerformanceParamsTemplate}>下载模板</Button><Button loading={performanceLoading} onClick={async () => {
                  setPerformanceLoading(true);
                  try {
                    const perf = ensureApiOk(await calculateValueScorePerformance());
                    setPerformanceResult((perf as any).data);
                    message.success('性能折算已刷新');
                  } catch (e) {
                    message.error(parseApiError(e, '刷新失败'));
                  } finally {
                    setPerformanceLoading(false);
                  }
                }}>刷新</Button></Space>}>
                  <Table rowKey="config_type" dataSource={performanceResult?.items || []} pagination={withTotalPagination(10)} columns={[
                    { title: '配置类型', dataIndex: 'config_type' },
                    { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores', render: (v: number) => formatInt(v) },
                    { title: '内存容量(GB)', dataIndex: 'memory_capacity_gb', render: (v: number) => formatFloat(v) },
                    { title: '不可用核数', dataIndex: 'unavailable_cores', render: (v: number) => formatInt(v) },
                    { title: '不可用内存(GB)', dataIndex: 'unavailable_memory_gb', render: (v: number) => formatFloat(v) },
                    { title: '性能跑分', dataIndex: 'performance_score', render: (v: number) => formatFloat(v) },
                    { title: '可用核数', dataIndex: 'available_cores', render: (v: number) => formatInt(v) },
                    { title: '可用内存(GB)', dataIndex: 'available_memory_gb', render: (v: number) => formatFloat(v) },
                    { title: '标准跑分', dataIndex: 'standard_score', render: (v: number) => formatFloat(v) },
                    { title: 'CPU性能折算系数', dataIndex: 'cpu_performance_factor', render: (v: number) => formatFloat(v) },
                    { title: '内存配比', dataIndex: 'memory_ratio', render: (v: number) => formatFloat(v) },
                    { title: '内存配比系数', dataIndex: 'memory_ratio_factor', render: (v: number) => formatFloat(v) },
                    { title: '整体性能折算比', dataIndex: 'overall_performance_ratio', render: (v: number) => formatFloat(v) }
                  ]} />
                  {performanceResult?.note ? <Text type="secondary">{performanceResult.note}</Text> : null}
                </Card>

                <Card title="服务器月TCO试算" extra={<Space><Button onClick={exportValueScoreTCO}>导出Excel</Button><Button loading={tcoLoading} onClick={async () => {
                  setTcoLoading(true);
                  try {
                    const tco = ensureApiOk(await calculateValueScoreTCO());
                    setTcoResult(tco.data);
                    message.success('月TCO已刷新');
                  } catch (e) {
                    message.error(parseApiError(e, '刷新失败'));
                  } finally {
                    setTcoLoading(false);
                  }
                }}>刷新试算</Button></Space>}>
                  <Table rowKey="config_type" dataSource={tcoResult?.items || []} pagination={withTotalPagination(10)} columns={[
                    { title: '配置类型', dataIndex: 'config_type' },
                    { title: '功率(W)', dataIndex: 'power_watts', render: (v: number) => formatFloat(v) },
                    { title: '功率(KW)', dataIndex: 'power_kw', render: (v: number) => formatFloat(v) },
                    { title: '机柜费/月', dataIndex: 'cabinet_cost_monthly', render: (v: number) => formatFloat(v) },
                    { title: '原值(CNY)', dataIndex: 'server_original_cny', render: (v: number) => formatFloat(v) },
                    { title: '折旧/月', dataIndex: 'depreciation_monthly', render: (v: number) => formatFloat(v) },
                    { title: '网络设备分摊/月', dataIndex: 'network_device_monthly', render: (v: number) => formatFloat(v) },
                    { title: '网络机柜等分摊/月', dataIndex: 'network_cabinet_monthly', render: (v: number) => formatFloat(v) },
                    { title: '服务器续保费/月', dataIndex: 'server_renewal_monthly', render: (v: number) => formatFloat(v) },
                    { title: '其他固定成本/月', dataIndex: 'other_fixed_cost_monthly', render: (v: number) => formatFloat(v) },
                    { title: '月TCO', dataIndex: 'total_tco_monthly', render: (v: number) => formatFloat(v) }
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
        ]}
      />
    </Space>
  );
}

// Keep existing helper implementations below unchanged.
type RegionKey = 'domestic' | 'india';
interface AssetSnapshotRow { region: '国内' | '印度'; snapshotLabel: string; snapshotDate: string; inWarranty: number; outWarranty: number; total: number; outWarrantyRatio: number; }
interface AssetTrendPoint { year: number; outCount: number; cumulativeOutCount: number; cumulativeOutRatio: number; }
function buildAssetAnalysis(servers: ServerItem[]) {
  const now = new Date();
  const nextYear0630 = new Date(now.getFullYear() + 1, 5, 30);
  const snapshots = [{ label: '当前时间', date: now }, { label: '次年6月30日', date: nextYear0630 }];
  const totals: Record<RegionKey, number> = { domestic: 0, india: 0 };
  const snapshotRows: AssetSnapshotRow[] = [];
  const trends: Record<RegionKey, AssetTrendPoint[]> = { domestic: [], india: [] };
  (['domestic', 'india'] as RegionKey[]).forEach((region) => {
    const list = servers.filter((s) => resolveRegion(s.idc) === region);
    totals[region] = list.length;
    for (const snap of snapshots) {
      let outWarranty = 0;
      for (const item of list) {
        const end = parseYMD(item.warranty_end_date);
        if (!end || end.getTime() < snap.date.getTime()) outWarranty += 1;
      }
      const total = list.length;
      snapshotRows.push({ region: region === 'domestic' ? '国内' : '印度', snapshotLabel: snap.label, snapshotDate: formatDate(snap.date), inWarranty: Math.max(0, total - outWarranty), outWarranty, total, outWarrantyRatio: total > 0 ? (outWarranty * 100) / total : 0 });
    }
    const yearMap = new Map<number, number>();
    for (const item of list) {
      const end = parseYMD(item.warranty_end_date);
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
  return { snapshotRows, trends, totals };
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
function formatInt(v?: number) { return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 }); }
function formatFloat(v?: number) { return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }); }
function formatMaybeNumber(v?: string) { const n = Number((v || '').trim()); if (Number.isNaN(n)) return v || '-'; return n.toLocaleString('en-US', { maximumFractionDigits: 2 }); }
