import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  DatePicker,
  Divider,
  Dropdown,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Slider,
  Space,
  Table,
  Tabs,
  Tooltip,
  Typography,
  Upload,
  message
} from 'antd';
import { ExclamationCircleOutlined, UploadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import type { UploadProps } from 'antd';
import { useNavigate } from 'react-router-dom';
import {
  createPlan,
  deletePlan,
  exportPlan,
  getRenewalSettings,
  importSpecialRules,
  listHostPackages,
  listPackageFailureRates,
  listPlans,
  listRenewalUnitPrices,
  listServers,
  listSpecialRules,
  listStorageTopServerRates,
  updateRenewalSettings,
  updateRenewalUnitPrices
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  HostPackageConfig,
  PackageFailureRate,
  RenewalPlan,
  RenewalPlanSettings,
  RenewalRequirements,
  RenewalTargetMode,
  RenewalUnitPrice,
  ServerItem,
  SpecialRule,
  StorageTopServerRate
} from '../types';

const { Text } = Typography;
const { RangePicker } = DatePicker;

const COUNTRY_OPTIONS = ['国内', '印度'] as const;
const SCENE_OPTIONS = [
  { key: 'compute', label: '计算' },
  { key: 'warm_storage', label: '温存储' },
  { key: 'hot_storage', label: '热存储' },
  { key: 'gpu', label: 'GPU' }
] as const;
const ALL_SCENE_OPTION_VALUE = '__ALL__';

export default function PlanPage() {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<RenewalPlanSettings>(defaultRenewalSettings());
  const [settingsEditing, setSettingsEditing] = useState(false);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [loading, setLoading] = useState(false);

  const [plans, setPlans] = useState<RenewalPlan[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [specialRules, setSpecialRules] = useState<SpecialRule[]>([]);
  const [specialLoading, setSpecialLoading] = useState(false);
  const [specialUploading, setSpecialUploading] = useState(false);
  const [activeTab, setActiveTab] = useState('plan');
  const [unitPrices, setUnitPrices] = useState<RenewalUnitPrice[]>([]);
  const [unitPriceLoading, setUnitPriceLoading] = useState(false);
  const [unitPriceSaving, setUnitPriceSaving] = useState(false);

  const [servers, setServers] = useState<ServerItem[]>([]);
  const [hostPackages, setHostPackages] = useState<HostPackageConfig[]>([]);
  const [packageFailureRates, setPackageFailureRates] = useState<PackageFailureRate[]>([]);
  const [storageTopServerRates, setStorageTopServerRates] = useState<StorageTopServerRate[]>([]);
  const [toolLoading, setToolLoading] = useState(false);

  const [toolCountry, setToolCountry] = useState<string>('国内');
  const [toolMode, setToolMode] = useState<'config_type' | 'sn'>('config_type');
  const [toolConfigSceneCategory, setToolConfigSceneCategory] = useState<string | undefined>(undefined);
  const [toolConfigType, setToolConfigType] = useState<string>('');
  const [toolSN, setToolSN] = useState<string>('');
  const [diskPrice, setDiskPrice] = useState<number>(0);
  const [logisticsCost, setLogisticsCost] = useState<number>(0);
  const [priceMultiplier, setPriceMultiplier] = useState<number>(1);

  const [queryPlanID, setQueryPlanID] = useState('');
  const [queryTargetDateRange, setQueryTargetDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null);
  const [queryExcludedPSA, setQueryExcludedPSA] = useState('');
  const [queryExcludedEnv, setQueryExcludedEnv] = useState('');

  async function loadInitialPlans() {
    setListLoading(true);
    try {
      const resp = ensureApiOk(await listPlans());
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载历史方案失败'));
    } finally {
      setListLoading(false);
    }
  }

  async function reloadSettings() {
    try {
      const resp = ensureApiOk(await getRenewalSettings());
      setSettings(normalizeSettings(resp.data));
    } catch (e) {
      message.error(parseApiError(e, '加载方案参数失败'));
    }
  }

  async function reloadPlans() {
    setListLoading(true);
    try {
      const params = {
        plan_id: queryPlanID.trim() || undefined,
        target_date_from: queryTargetDateRange?.[0]?.format('YYYY-MM-DD') || undefined,
        target_date_to: queryTargetDateRange?.[1]?.format('YYYY-MM-DD') || undefined,
        excluded_psa: queryExcludedPSA.trim() || undefined,
        excluded_environment: queryExcludedEnv.trim() || undefined
      };
      const resp = ensureApiOk(await listPlans(params));
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载历史方案失败'));
    } finally {
      setListLoading(false);
    }
  }

  async function reloadSpecialRules() {
    setSpecialLoading(true);
    try {
      const resp = ensureApiOk(await listSpecialRules());
      setSpecialRules(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载例外清单失败'));
    } finally {
      setSpecialLoading(false);
    }
  }

  async function reloadUnitPrices() {
    setUnitPriceLoading(true);
    try {
      const resp = ensureApiOk(await listRenewalUnitPrices());
      setUnitPrices(normalizeUnitPrices(resp.data.list || []));
    } catch (e) {
      message.error(parseApiError(e, '加载续保单价失败'));
    } finally {
      setUnitPriceLoading(false);
    }
  }

  async function reloadToolBaseData() {
    setToolLoading(true);
    try {
      const [serverResp, hostPackageResp, packageRateResp, warmTopResp, hotTopResp] = await Promise.all([
        listServers(),
        listHostPackages(),
        listPackageFailureRates(),
        listStorageTopServerRates('warm_storage'),
        listStorageTopServerRates('hot_storage')
      ]);
      setServers(ensureApiOk(serverResp).data.list || []);
      setHostPackages(ensureApiOk(hostPackageResp).data.list || []);
      setPackageFailureRates(ensureApiOk(packageRateResp).data.list || []);
      const warmTop = ensureApiOk(warmTopResp).data.list || [];
      const hotTop = ensureApiOk(hotTopResp).data.list || [];
      setStorageTopServerRates([...warmTop, ...hotTop]);
    } catch (e) {
      message.error(parseApiError(e, '加载续保/自维修分析基础数据失败'));
    } finally {
      setToolLoading(false);
    }
  }

  useEffect(() => {
    loadInitialPlans();
    reloadSpecialRules();
    reloadUnitPrices();
    reloadToolBaseData();
    reloadSettings();
  }, []);

  async function onCreatePlan() {
    if (!settings.target_date) {
      message.warning('请选择续保目标时间');
      return;
    }
    if (settings.domestic_budget < 0 || settings.india_budget < 0) {
      message.warning('预算不能为负数');
      return;
    }

    setLoading(true);
    try {
      const resp = ensureApiOk(await createPlan({
        target_date: settings.target_date,
        excluded_environments: settings.excluded_environments,
        excluded_psas: settings.excluded_psas,
        requirements: settings.requirements,
        domestic_budget: settings.domestic_budget,
        india_budget: settings.india_budget
      }));
      message.success(`方案已生成：${resp.data.plan_id}`);
      await reloadPlans();
      navigate(`/plan/${resp.data.plan_id}`);
    } catch (e) {
      message.error(parseApiError(e, '生成失败'));
    } finally {
      setLoading(false);
    }
  }

  async function onDelete(planId: string) {
    try {
      ensureApiOk(await deletePlan(planId));
      message.success('方案已删除');
      await reloadPlans();
    } catch (e) {
      message.error(parseApiError(e, '删除失败'));
    }
  }

  async function onResetHistoryFilters() {
    setQueryPlanID('');
    setQueryTargetDateRange(null);
    setQueryExcludedPSA('');
    setQueryExcludedEnv('');
    setListLoading(true);
    try {
      const resp = ensureApiOk(await listPlans());
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '重置查询失败'));
    } finally {
      setListLoading(false);
    }
  }

  function onUnitPriceChange(country: string, scene: string, next: number) {
    setUnitPrices((prev) => prev.map((x) => (
      x.country === country && x.scene_category === scene
        ? { ...x, unit_price: next }
        : x
    )));
  }

  async function onSaveUnitPrices() {
    setUnitPriceSaving(true);
    try {
      const payload = normalizeUnitPrices(unitPrices);
      const resp = ensureApiOk(await updateRenewalUnitPrices(payload));
      setUnitPrices(normalizeUnitPrices(resp.data.list || []));
      message.success('续保单价已保存');
    } catch (e) {
      message.error(parseApiError(e, '保存续保单价失败'));
    } finally {
      setUnitPriceSaving(false);
    }
  }

  function onSettingListChange(field: 'excluded_environments' | 'excluded_psas', raw: string) {
    const values = raw.split(/[，,]/).map((x) => x.trim()).filter(Boolean);
    setSettings((prev) => ({ ...prev, [field]: values }));
  }

  function onSettingTarget(region: 'domestic' | 'india', scene: keyof RenewalRequirements['domestic'], patch: Partial<{ mode: RenewalTargetMode; target: number }>) {
    setSettings((prev) => ({
      ...prev,
      requirements: {
        ...prev.requirements,
        [region]: {
          ...prev.requirements[region],
          [scene]: {
            ...prev.requirements[region][scene],
            ...patch
          }
        }
      }
    }));
  }

  async function onSaveSettings() {
    setSettingsSaving(true);
    try {
      const resp = ensureApiOk(await updateRenewalSettings(settings));
      setSettings(normalizeSettings(resp.data));
      setSettingsEditing(false);
      message.success('方案参数已保存');
    } catch (e) {
      message.error(parseApiError(e, '保存方案参数失败'));
    } finally {
      setSettingsSaving(false);
    }
  }

  const specialUploadProps: UploadProps = {
    maxCount: 1,
    showUploadList: true,
    accept: '.xlsx',
    customRequest: async (options) => {
      const file = options.file as File;
      setSpecialUploading(true);
      try {
        const resp = ensureApiOk(await importSpecialRules(file));
        message.success(`例外清单导入完成：成功 ${resp.data.success} 条`);
        await reloadSpecialRules();
        options.onSuccess?.({}, new XMLHttpRequest());
      } catch (e) {
        message.error(parseApiError(e, '导入例外清单失败'));
        options.onError?.(new Error('import failed'));
      } finally {
        setSpecialUploading(false);
      }
    }
  };

  const columns = [
    { title: '方案ID', dataIndex: 'plan_id', width: 140 },
    { title: '续保目标时间', dataIndex: 'target_date', width: 120 },
    {
      title: '排除PSA',
      dataIndex: 'excluded_psas',
      render: (v: string[]) => (v && v.length ? v.join('、') : '-'),
      width: 180
    },
    {
      title: '排除环境',
      dataIndex: 'excluded_environments',
      render: (v: string[]) => (v && v.length ? v.join('、') : '-'),
      width: 160
    },
    { title: '计算目标核数', dataIndex: 'target_cores', width: 120, render: (v: number) => formatInt(v) },
    { title: '温存储空间需求(TB)', dataIndex: 'warm_target_storage_tb', width: 160, render: (v: number) => formatFloat(v) },
    { title: '热存储空间需求(TB)', dataIndex: 'hot_target_storage_tb', width: 160, render: (v: number) => formatFloat(v) },
    {
      title: '摘要',
      width: 260,
      render: (_: unknown, r: RenewalPlan) => {
        const summary = buildPlanSummary(r);
        return (
          <Space direction="vertical" size={2}>
            <Text type={summary.computeRate >= 100 ? undefined : 'warning'}>算力达成: {summary.computeRate.toFixed(1)}%</Text>
            <Text type={summary.warmRate >= 100 ? undefined : 'warning'}>温存储达成: {summary.warmRate.toFixed(1)}%</Text>
            <Text type={summary.hotRate >= 100 ? undefined : 'warning'}>热存储达成: {summary.hotRate.toFixed(1)}%</Text>
            <Text type="secondary">入选台数: {formatInt(r.selected_count || 0)}</Text>
          </Space>
        );
      }
    },
    {
      title: '异常',
      width: 90,
      render: (_: unknown, r: RenewalPlan) => {
        const anomaly = analyzeAnomalies(r);
        const report = buildAnomalyReport(r);
        return (
          <Tooltip title={report}>
            <Button
              type={anomaly.blockers.length ? 'primary' : 'default'}
              danger={anomaly.blockers.length > 0}
              icon={<ExclamationCircleOutlined />}
              size="small"
              onClick={() => Modal.info({ title: `方案 ${r.plan_id} 异常报告`, width: 760, content: <pre style={{ whiteSpace: 'pre-wrap' }}>{report}</pre> })}
            >
              {anomaly.blockers.length + anomaly.warnings.length || 0}
            </Button>
          </Tooltip>
        );
      }
    },
    {
      title: '操作',
      fixed: 'right' as const,
      width: 280,
      render: (_: unknown, r: RenewalPlan) => (
        <Space>
          <Button size="small" onClick={() => navigate(`/plan/${r.plan_id}`)}>查看</Button>
          <Dropdown
            menu={{
              items: [
                { key: 'xlsx', label: '下载 Excel (.xlsx)' },
                { key: 'csv', label: '下载 CSV (.csv)' }
              ],
              onClick: ({ key }) => exportPlan(r.plan_id, key as 'xlsx' | 'csv')
            }}
          >
            <Button size="small">下载</Button>
          </Dropdown>
          <Popconfirm title="确认删除该方案？" onConfirm={() => onDelete(r.plan_id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const serverBySN = useMemo(() => {
    const map = new Map<string, ServerItem>();
    servers.forEach((x) => map.set(x.sn, x));
    return map;
  }, [servers]);

  const hostPackageByConfigType = useMemo(() => {
    const map = new Map<string, HostPackageConfig>();
    hostPackages.forEach((x) => {
      const key = normalizeConfigType(x.config_type);
      if (key) {
        map.set(key, x);
      }
    });
    return map;
  }, [hostPackages]);

  const afrByConfigType = useMemo(() => {
    const map = new Map<string, number>();
    packageFailureRates.forEach((x) => {
      const afr = Number(x.recent_1y_failure_rate ?? x.failure_rate ?? 0);
      const key = normalizeConfigType(x.config_type);
      if (key) {
        map.set(key, afr);
      }
    });
    return map;
  }, [packageFailureRates]);

  const afrBySN = useMemo(() => {
    const map = new Map<string, number>();
    storageTopServerRates.forEach((x) => {
      const sn = String(x.sn || '').trim();
      if (!sn) return;
      map.set(sn, Number(x.fault_rate || 0));
    });
    return map;
  }, [storageTopServerRates]);

  const configTypeToSceneCategory = useMemo(() => {
    const map = new Map<string, string>();
    hostPackages.forEach((x) => {
      const key = normalizeConfigType(x.config_type);
      if (!key) return;
      map.set(key, normalizeSceneCategory(x.scene_category));
    });
    return map;
  }, [hostPackages]);

  const configTypeOptions = useMemo(() => {
    const set = new Set<string>();
    hostPackages.forEach((x) => { if (x.config_type) set.add(x.config_type); });
    servers.forEach((x) => { if (x.config_type) set.add(x.config_type); });
    return Array.from(set).sort().map((x) => ({ label: x, value: x }));
  }, [hostPackages, servers]);

  const filteredConfigTypeOptions = useMemo(() => {
    if (!toolConfigSceneCategory) {
      return configTypeOptions;
    }
    return configTypeOptions.filter((x) => (
      configTypeToSceneCategory.get(normalizeConfigType(x.value)) === toolConfigSceneCategory
    ));
  }, [configTypeOptions, configTypeToSceneCategory, toolConfigSceneCategory]);

  const snOptions = useMemo(() => (
    servers
      .filter((x) => x.sn)
      .map((x) => ({
        label: `${x.sn}${x.config_type ? `（${x.config_type}）` : ''}`,
        value: x.sn
      }))
      .sort((a, b) => a.value.localeCompare(b.value))
  ), [servers]);

  useEffect(() => {
    if (toolMode !== 'config_type' || !toolConfigType) return;
    const stillAvailable = filteredConfigTypeOptions.some((x) => x.value === toolConfigType);
    if (!stillAvailable) {
      setToolConfigType('');
    }
  }, [toolMode, toolConfigType, filteredConfigTypeOptions]);

  const selectedConfigTypeRaw = toolMode === 'sn'
    ? (toolSN ? serverBySN.get(toolSN)?.config_type || '' : '')
    : toolConfigType;
  const selectedConfigType = normalizeConfigType(selectedConfigTypeRaw);

  const selectedHostPackage = selectedConfigType ? hostPackageByConfigType.get(selectedConfigType) : undefined;
  const selectedConfigTypeLabel = selectedHostPackage?.config_type || selectedConfigTypeRaw || '-';
  const sceneCategoryRaw = selectedHostPackage?.scene_category || 'warm_storage';
  const sceneCategory = normalizeSceneCategory(sceneCategoryRaw);
  const diskCount = Number(selectedHostPackage?.data_disk_count || 0);
  const hasConfigTypeAfr = Boolean(selectedConfigType && afrByConfigType.has(selectedConfigType));
  const configTypeAfr = Number((selectedConfigType ? afrByConfigType.get(selectedConfigType) : undefined) ?? 0);
  const hasDeviceAfr = Boolean(toolMode === 'sn' && toolSN && afrBySN.has(toolSN));
  const selectedDeviceAfr = Number((toolMode === 'sn' && toolSN ? afrBySN.get(toolSN) : undefined) ?? 0);
  const effectiveAfr = toolMode === 'sn' ? selectedDeviceAfr : configTypeAfr;
  const configTypeAfrDisplay = hasConfigTypeAfr ? formatPercent(configTypeAfr) : '-';
  const deviceAfrDisplay = toolMode === 'sn'
    ? (toolSN ? (hasDeviceAfr ? formatPercent(selectedDeviceAfr) : '-') : '-')
    : '-';

  const countryOptions = useMemo(() => {
    const set = new Set<string>(COUNTRY_OPTIONS as unknown as string[]);
    unitPrices.forEach((x) => {
      const v = String(x.country || '').trim();
      if (v) set.add(v);
    });
    return Array.from(set).map((x) => ({ label: x, value: x }));
  }, [unitPrices]);

  const matchedUnitPrice = unitPrices.find((x) => (
    normalizeCountry(x.country) === normalizeCountry(toolCountry)
    && normalizeSceneCategory(x.scene_category) === sceneCategory
  ));
  const baseRenewalPrice = Number(matchedUnitPrice?.unit_price || 0);
  const simulatedRenewalPrice = baseRenewalPrice * priceMultiplier;
  const denominator = diskCount + 1;
  const renewalCostPerDisk = denominator > 0 ? simulatedRenewalPrice / denominator : 0;
  const selfRepairCost = (Number(diskPrice || 0) + Number(logisticsCost || 0)) * effectiveAfr;

  const toolWarnings: string[] = [];
  if (!selectedConfigType) {
    toolWarnings.push('请先选择配置类型或 SN。');
  }
  if (selectedConfigType && !selectedHostPackage) {
    toolWarnings.push(`配置类型 ${selectedConfigTypeLabel} 在主机套餐表中无记录，无法获取磁盘数量。`);
  }
  if (toolMode === 'config_type' && selectedConfigType && !afrByConfigType.has(selectedConfigType)) {
    toolWarnings.push(`配置类型 ${selectedConfigTypeLabel} 未找到 AFR（套型故障率）数据。`);
  }
  if (toolMode === 'sn' && toolSN && !afrBySN.has(toolSN)) {
    toolWarnings.push(`SN ${toolSN} 未找到设备 AFR（温/热存储故障 TOP100 同源数据）。`);
  }
  if (selectedConfigType && !matchedUnitPrice) {
    toolWarnings.push(`未找到国家=${toolCountry}、场景=${sceneCategoryLabel(sceneCategory)} 的续保单价。`);
  }

  const decision = selectedConfigType
    ? renewalCostPerDisk <= selfRepairCost
      ? '建议续保（单盘续保成本更低）'
      : '建议买新盘自维修（期望成本更低）'
    : '待分析（请先选择配置类型或 SN）';
  const decisionType: 'success' | 'warning' | 'info' = selectedConfigType ? (renewalCostPerDisk <= selfRepairCost ? 'success' : 'warning') : 'info';

  const costGap = renewalCostPerDisk - selfRepairCost;
  const absCostGap = Math.abs(costGap);
  const gapRatio = selfRepairCost !== 0 ? costGap / selfRepairCost : 0;
  const gapDirection = selectedConfigType ? (costGap <= 0 ? '续保更省' : '自维修更省') : '-';

  function onExportToolCSV() {
    if (!selectedConfigType) {
      message.warning('请先选择配置类型或 SN 再导出。');
      return;
    }

    const headers = [
      '分析时间', '国家', '分析维度', '配置类型', 'SN', '场景', '磁盘数量', '配置类型AFR', '设备AFR', '生效AFR',
      '单盘价格', '物流成本', '续保单价基准', '涨价倍数', '续保单价模拟',
      '单盘续保成本', '买新盘自维修成本', '价差(续保-自维修)', '价差比例(相对自维修)', '结论', '备注'
    ];

    const row = [
      dayjs().format('YYYY-MM-DD HH:mm:ss'),
      toolCountry,
      toolMode === 'sn' ? 'SN' : '配置类型',
      selectedConfigTypeLabel === '-' ? '' : selectedConfigTypeLabel,
      toolMode === 'sn' ? toolSN : '',
      sceneCategoryLabel(sceneCategory),
      String(diskCount),
      configTypeAfrDisplay,
      deviceAfrDisplay,
      formatPercent(effectiveAfr),
      formatMoney(diskPrice),
      formatMoney(logisticsCost),
      formatMoney(baseRenewalPrice),
      priceMultiplier.toFixed(1),
      formatMoney(simulatedRenewalPrice),
      formatMoney(renewalCostPerDisk),
      formatMoney(selfRepairCost),
      formatMoney(costGap),
      formatSignedPercent(gapRatio),
      decision,
      toolWarnings.join('；')
    ];

    const csv = [headers, row]
      .map((line) => line.map((x) => `"${String(x).replace(/"/g, '""')}"`).join(','))
      .join('\n');

    const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `renewal-self-repair-analysis-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  return (
    <Tabs
      activeKey={activeTab}
      onChange={setActiveTab}
      items={[
        {
          key: 'plan',
          label: '续保方案管理',
          children: (
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <Card
                title="续保管理 - 生成方案"
                extra={(
                  <Space>
                    {!settingsEditing && <Button onClick={() => setSettingsEditing(true)}>修改方案参数</Button>}
                    {settingsEditing && (
                      <>
                        <Button onClick={() => { setSettingsEditing(false); reloadSettings(); }}>取消</Button>
                        <Button type="primary" loading={settingsSaving} onClick={onSaveSettings}>保存</Button>
                      </>
                    )}
                    <Button type="primary" loading={loading} onClick={onCreatePlan}>生成方案</Button>
                  </Space>
                )}
              >
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                  <Space wrap>
                    <Text>续保目标时间</Text>
                    <DatePicker
                      value={settings.target_date ? dayjs(settings.target_date) : undefined}
                      onChange={(d) => setSettings((prev) => ({ ...prev, target_date: d ? d.format('YYYY-MM-DD') : '' }))}
                      allowClear={false}
                      disabled={!settingsEditing}
                    />
                  </Space>

                  <Space wrap>
                    <Text>排除环境</Text>
                    {settingsEditing ? (
                      <Input style={{ width: 300 }} value={(settings.excluded_environments || []).join(',')} onChange={(e) => onSettingListChange('excluded_environments', e.target.value)} placeholder="开发,测试" />
                    ) : (
                      <Text>{(settings.excluded_environments || []).join('、') || '-'}</Text>
                    )}
                  </Space>

                  <Space wrap>
                    <Text>排除PSA</Text>
                    {settingsEditing ? (
                      <Input style={{ width: 300 }} value={(settings.excluded_psas || []).join(',')} onChange={(e) => onSettingListChange('excluded_psas', e.target.value)} placeholder="例如：A,B,C 或 10,20" />
                    ) : (
                      <Text>{(settings.excluded_psas || []).join('、') || '-'}</Text>
                    )}
                  </Space>

                  <Table
                    pagination={false}
                    rowKey={(r) => `${r.region}-${r.scene}`}
                    dataSource={buildRequirementRows(settings.requirements)}
                    columns={[
                      { title: '国家', dataIndex: 'regionLabel', width: 100 },
                      { title: '场景', dataIndex: 'sceneLabel', width: 120 },
                      {
                        title: '需求方式',
                        dataIndex: 'mode',
                        width: 140,
                        render: (v: RenewalTargetMode, row: any) => settingsEditing ? (
                          <Select
                            style={{ width: 120 }}
                            value={v}
                            options={[{ label: '手动输入', value: 'manual' }, { label: '多多益善', value: 'maximize' }]}
                            onChange={(next) => onSettingTarget(row.region, row.scene, { mode: next as RenewalTargetMode })}
                          />
                        ) : (v === 'maximize' ? '多多益善' : '手动输入')
                      },
                      {
                        title: '需求值',
                        dataIndex: 'target',
                        width: 180,
                        render: (v: number, row: any) => {
                          if (!settingsEditing) {
                            return row.mode === 'maximize' ? '价值分全部达标' : formatInt(v);
                          }
                          return (
                            <InputNumber
                              min={0}
                              step={row.scene === 'gpu' ? 1 : 10}
                              value={Number(v || 0)}
                              disabled={row.mode === 'maximize'}
                              onChange={(next) => onSettingTarget(row.region, row.scene, { target: Number(next ?? 0) })}
                            />
                          );
                        }
                      }
                    ]}
                  />

                  <Space wrap>
                    <Text>国内预算(CNY)：{settingsEditing ? '' : formatInt(settings.domestic_budget)}</Text>
                    {settingsEditing && <InputNumber min={0} step={1000} value={settings.domestic_budget} onChange={(v) => setSettings((prev) => ({ ...prev, domestic_budget: Number(v ?? 0) }))} />}
                    <Text>印度预算(CNY)：{settingsEditing ? '' : formatInt(settings.india_budget)}</Text>
                    {settingsEditing && <InputNumber min={0} step={1000} value={settings.india_budget} onChange={(v) => setSettings((prev) => ({ ...prev, india_budget: Number(v ?? 0) }))} />}
                  </Space>
                </Space>
              </Card>

              <Card
                title="续保管理 - 历史方案列表"
                extra={(
                  <Space wrap>
                    <Input
                      style={{ width: 160 }}
                      value={queryPlanID}
                      onChange={(e) => setQueryPlanID(e.target.value)}
                      placeholder="方案ID"
                      allowClear
                    />
                    <RangePicker
                      value={queryTargetDateRange}
                      onChange={(v) => setQueryTargetDateRange((v as [dayjs.Dayjs | null, dayjs.Dayjs | null]) || null)}
                      allowEmpty={[true, true]}
                    />
                    <Input
                      style={{ width: 140 }}
                      value={queryExcludedPSA}
                      onChange={(e) => setQueryExcludedPSA(e.target.value)}
                      placeholder="排除PSA"
                      allowClear
                    />
                    <Input
                      style={{ width: 140 }}
                      value={queryExcludedEnv}
                      onChange={(e) => setQueryExcludedEnv(e.target.value)}
                      placeholder="排除环境"
                      allowClear
                    />
                    <Button type="primary" onClick={reloadPlans} loading={listLoading}>搜索</Button>
                    <Button onClick={onResetHistoryFilters} loading={listLoading}>重置</Button>
                  </Space>
                )}
              >
                <Table
                  rowKey="plan_id"
                  loading={listLoading}
                  dataSource={plans}
                  columns={columns}
                  scroll={{ x: 1500 }}
                  pagination={withTotalPagination(10)}
                />
              </Card>

            </Space>
          )
        },
        {
          key: 'special_rules',
          label: '例外清单维护',
          children: (
            <Card
              title="续保管理 - 例外清单"
              extra={(
                <Space>
                  <Button onClick={reloadSpecialRules} loading={specialLoading}>刷新</Button>
                  <Upload {...specialUploadProps}>
                    <Button icon={<UploadOutlined />} loading={specialUploading}>上传并导入</Button>
                  </Upload>
                </Space>
              )}
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                <Text type="secondary">导入模板建议三列：SN、策略（加白/加黑）、原因（可选）。其余字段会自动从服务器管理表按 SN 补全。</Text>
                <Table
                  rowKey="sn"
                  loading={specialLoading}
                  dataSource={specialRules}
                  pagination={withTotalPagination(10)}
                  columns={[
                    { title: 'SN', dataIndex: 'sn', width: 160 },
                    { title: '制造商', dataIndex: 'manufacturer', width: 120 },
                    { title: '型号', dataIndex: 'model', width: 140 },
                    { title: 'PSA', dataIndex: 'psa', width: 100 },
                    { title: '套餐', dataIndex: 'package_type', width: 140 },
                    { title: '策略', dataIndex: 'policy', width: 100 },
                    { title: '原因', dataIndex: 'reason', width: 220 }
                  ]}
                  scroll={{ x: 1140 }}
                />
              </Space>
            </Card>
          )
        },
        {
          key: 'unit_price',
          label: '续保单价维护',
          children: (
            <Card
              title="续保单价维护（国家 + 场景大类）"
              extra={(
                <Space>
                  <Button onClick={reloadUnitPrices} loading={unitPriceLoading}>刷新</Button>
                  <Button type="primary" onClick={onSaveUnitPrices} loading={unitPriceSaving}>保存</Button>
                </Space>
              )}
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                <Text type="secondary">当前支持国家：国内、印度；场景大类：计算、温存储、热存储、GPU。修改后点击保存，单价会入库并用于续保金额估算。</Text>
                <Table
                  rowKey={(r) => `${r.country}-${r.scene_category}`}
                  loading={unitPriceLoading}
                  pagination={false}
                  dataSource={unitPrices}
                  columns={[
                    { title: '国家', dataIndex: 'country', width: 120 },
                    {
                      title: '场景大类',
                      dataIndex: 'scene_category',
                      width: 120,
                      render: (v: string) => SCENE_OPTIONS.find((x) => x.key === v)?.label || v
                    },
                    {
                      title: '续保单价',
                      dataIndex: 'unit_price',
                      width: 200,
                      render: (v: number, row: RenewalUnitPrice) => (
                        <InputNumber
                          min={0}
                          precision={2}
                          step={10}
                          style={{ width: 180 }}
                          value={v}
                          onChange={(next) => onUnitPriceChange(row.country, row.scene_category, Number(next ?? 0))}
                        />
                      )
                    }
                  ]}
                />
              </Space>
            </Card>
          )
        },
        {
          key: 'renewal_tool',
          label: '续保/自维修分析',
          children: (
            <Card
              title="续保管理 - 小工具（续保 vs 买新盘自维修）"
              extra={(
                <Space>
                  <Button onClick={reloadToolBaseData} loading={toolLoading}>刷新基础数据</Button>
                  <Button onClick={onExportToolCSV}>导出当前分析CSV</Button>
                </Space>
              )}
            >
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Text type="secondary">
                  核心逻辑：续保成本 = 续保价格 / (磁盘数量 + 1)；买新盘自维修成本 = (单盘价格 + 物流成本) × AFR。分析对象=配置类型时使用配置类型AFR；分析对象=SN时使用设备AFR（来自温/热存储故障TOP100同源数据）。
                </Text>

                <Space wrap>
                  <Text>国家</Text>
                  <Select
                    style={{ width: 140 }}
                    value={toolCountry}
                    options={countryOptions}
                    onChange={setToolCountry}
                  />

                  <Text>分析对象</Text>
                  <Select
                    style={{ width: 160 }}
                    value={toolMode}
                    options={[{ label: '配置类型', value: 'config_type' }, { label: 'SN', value: 'sn' }]}
                    onChange={(v) => setToolMode(v as 'config_type' | 'sn')}
                  />

                  {toolMode === 'config_type' ? (
                    <>
                      <Text>配置大类</Text>
                      <Select
                        allowClear
                        style={{ width: 160 }}
                        value={toolConfigSceneCategory || ALL_SCENE_OPTION_VALUE}
                        options={[
                          { label: '全部', value: ALL_SCENE_OPTION_VALUE },
                          ...SCENE_OPTIONS.map((x) => ({ label: x.label, value: x.key }))
                        ]}
                        placeholder="全部"
                        onChange={(v) => setToolConfigSceneCategory(v === ALL_SCENE_OPTION_VALUE ? undefined : v)}
                      />
                      <Select
                        showSearch
                        optionFilterProp="label"
                        style={{ width: 360 }}
                        value={toolConfigType || undefined}
                        options={filteredConfigTypeOptions}
                        placeholder="选择配置类型"
                        onChange={setToolConfigType}
                      />
                    </>
                  ) : (
                    <Select
                      showSearch
                      optionFilterProp="label"
                      style={{ width: 420 }}
                      value={toolSN || undefined}
                      options={snOptions}
                      placeholder="选择 SN"
                      onChange={setToolSN}
                    />
                  )}
                </Space>

                <Space wrap>
                  <Text>单盘价格</Text>
                  <InputNumber
                    min={0}
                    precision={2}
                    step={100}
                    style={{ width: 180 }}
                    value={diskPrice}
                    onChange={(v) => setDiskPrice(Number(v ?? 0))}
                    addonAfter="CNY"
                  />

                  <Text>物流成本</Text>
                  <InputNumber
                    min={0}
                    precision={2}
                    step={50}
                    style={{ width: 180 }}
                    value={logisticsCost}
                    onChange={(v) => setLogisticsCost(Number(v ?? 0))}
                    addonAfter="CNY"
                  />
                </Space>

                <div>
                  <Text>续保涨价模拟（天平）：{priceMultiplier.toFixed(1)}x</Text>
                  <Slider
                    min={0.1}
                    max={10}
                    step={0.1}
                    tooltip={{ formatter: (v) => `${Number(v || 0).toFixed(1)}x` }}
                    value={priceMultiplier}
                    onChange={(v) => setPriceMultiplier(Number(v))}
                  />
                </div>

                <Divider style={{ margin: '8px 0' }} />

                <Space wrap size="large">
                  <Text>配置类型：<Text strong>{selectedConfigTypeLabel}</Text></Text>
                  <Text>场景：<Text strong>{sceneCategoryLabel(sceneCategory)}</Text></Text>
                  <Text>磁盘数量：<Text strong>{formatInt(diskCount)}</Text></Text>
                  <Text>配置类型AFR：<Text strong>{configTypeAfrDisplay}</Text></Text>
                  <Text type={toolMode === 'sn' ? undefined : 'secondary'}>
                    设备AFR：<Text strong>{deviceAfrDisplay}</Text>
                  </Text>
                </Space>

                <Space wrap size="large">
                  <Text>续保单价（基准）：<Text strong>{formatMoney(baseRenewalPrice)}</Text></Text>
                  <Text>续保单价（模拟后）：<Text strong>{formatMoney(simulatedRenewalPrice)}</Text></Text>
                  <Text>单盘续保成本：<Text strong>{formatMoney(renewalCostPerDisk)}</Text></Text>
                  <Text>买新盘自维修成本：<Text strong>{formatMoney(selfRepairCost)}</Text></Text>
                  <Text>价差（续保-自维修）：<Text strong>{formatMoney(costGap)}</Text></Text>
                  <Text>价差比例：<Text strong>{formatSignedPercent(gapRatio)}</Text></Text>
                </Space>

                <Alert
                  type={decisionType}
                  message={decision}
                  description={selectedConfigType
                    ? `结论依据：${formatMoney(renewalCostPerDisk)} vs ${formatMoney(selfRepairCost)}（取更低）；${gapDirection}，绝对价差 ${formatMoney(absCostGap)}，比例 ${formatSignedPercent(gapRatio)}`
                    : '请选择分析对象后自动给出结论'}
                  showIcon
                />

                {toolWarnings.length > 0 && (
                  <Alert
                    type="info"
                    showIcon
                    message="数据提示"
                    description={(
                      <ul style={{ margin: 0, paddingLeft: 20 }}>
                        {toolWarnings.map((x) => <li key={x}>{x}</li>)}
                      </ul>
                    )}
                  />
                )}
              </Space>
            </Card>
          )
        }
      ]}
    />
  );
}

function defaultRenewalSettings(): RenewalPlanSettings {
  return {
    target_date: dayjs().format('YYYY-MM-DD'),
    excluded_environments: ['开发', '测试'],
    excluded_psas: [],
    requirements: {
      domestic: {
        compute: { mode: 'manual', target: 1200 },
        warm_storage: { mode: 'manual', target: 0 },
        hot_storage: { mode: 'manual', target: 0 },
        gpu: { mode: 'manual', target: 0 }
      },
      india: {
        compute: { mode: 'manual', target: 0 },
        warm_storage: { mode: 'manual', target: 0 },
        hot_storage: { mode: 'manual', target: 0 },
        gpu: { mode: 'manual', target: 0 }
      }
    },
    domestic_budget: 0,
    india_budget: 0
  };
}

function normalizeSettings(input?: RenewalPlanSettings): RenewalPlanSettings {
  const base = defaultRenewalSettings();
  if (!input) return base;
  return {
    ...base,
    ...input,
    requirements: {
      domestic: { ...base.requirements.domestic, ...(input.requirements?.domestic || {}) },
      india: { ...base.requirements.india, ...(input.requirements?.india || {}) }
    }
  };
}

function buildRequirementRows(requirements: RenewalRequirements) {
  return [
    { region: 'domestic', regionLabel: '国内', scene: 'compute', sceneLabel: '计算', ...requirements.domestic.compute },
    { region: 'domestic', regionLabel: '国内', scene: 'warm_storage', sceneLabel: '温存储', ...requirements.domestic.warm_storage },
    { region: 'domestic', regionLabel: '国内', scene: 'hot_storage', sceneLabel: '热存储', ...requirements.domestic.hot_storage },
    { region: 'domestic', regionLabel: '国内', scene: 'gpu', sceneLabel: 'GPU', ...requirements.domestic.gpu },
    { region: 'india', regionLabel: '印度', scene: 'compute', sceneLabel: '计算', ...requirements.india.compute },
    { region: 'india', regionLabel: '印度', scene: 'warm_storage', sceneLabel: '温存储', ...requirements.india.warm_storage },
    { region: 'india', regionLabel: '印度', scene: 'hot_storage', sceneLabel: '热存储', ...requirements.india.hot_storage },
    { region: 'india', regionLabel: '印度', scene: 'gpu', sceneLabel: 'GPU', ...requirements.india.gpu }
  ];
}

function normalizeUnitPrices(list: RenewalUnitPrice[]): RenewalUnitPrice[] {
  const byKey = new Map(list.map((x) => [`${x.country}|${x.scene_category}`, { ...x, unit_price: Number(x.unit_price || 0) }]));
  const out: RenewalUnitPrice[] = [];
  for (const country of COUNTRY_OPTIONS) {
    for (const scene of SCENE_OPTIONS) {
      out.push(byKey.get(`${country}|${scene.key}`) || { country, scene_category: scene.key, unit_price: 0 });
    }
  }
  return out;
}

function withTotalPagination(pageSize: number) {
  return {
    pageSize,
    showTotal: (total: number) => `共${total}条，${Math.ceil(total / pageSize)}页`
  };
}

function buildAnomalyReport(plan: RenewalPlan): string {
  const detail = analyzeAnomalies(plan);
  const lines: string[] = [];

  lines.push('【阻断】');
  if (detail.blockers.length) {
    detail.blockers.forEach((x) => lines.push(`- ${x}`));
  } else {
    lines.push('- 无');
  }

  lines.push('');
  lines.push('【警告】');
  if (detail.warnings.length) {
    detail.warnings.forEach((x) => lines.push(`- ${x}`));
  } else {
    lines.push('- 无');
  }

  lines.push('');
  lines.push(`【结论】阻断 ${detail.blockers.length} 项，警告 ${detail.warnings.length} 项`);
  return lines.join('\n');
}

function analyzeAnomalies(plan: RenewalPlan): { blockers: string[]; warnings: string[] } {
  const blockers: string[] = [];
  const warnings: string[] = [];

  if ((plan.unmatched_config_count || 0) > 0) {
    blockers.push(`未匹配套餐配置: ${plan.unmatched_config_count} 个（${(plan.unmatched_config_types || []).join('、')}）`);
  }

  const computeSelected = plan.sections?.find((s) => s.bucket === 'compute')?.selected_cores || 0;
  if ((plan.required_compute_cores || 0) > computeSelected) {
    blockers.push(`计算型缺口未补齐（要求 ${plan.required_compute_cores || 0}，已选 ${computeSelected}）`);
  }

  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  if ((warm?.required_storage_tb || 0) > (warm?.selected_storage_tb || 0)) {
    blockers.push(`温存储缺口未补齐（要求 ${warm?.required_storage_tb || 0}TB，已选 ${warm?.selected_storage_tb || 0}TB）`);
  }

  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');
  if ((hot?.required_storage_tb || 0) > (hot?.selected_storage_tb || 0)) {
    blockers.push(`热存储缺口未补齐（要求 ${hot?.required_storage_tb || 0}TB，已选 ${hot?.selected_storage_tb || 0}TB）`);
  }

  if (!plan.items?.length) {
    warnings.push('方案列表为空，请检查输入与过滤条件');
  }

  return { blockers, warnings };
}

function buildPlanSummary(plan: RenewalPlan) {
  const computeRate = calcRate(plan.required_compute_cores || 0, plan.sections?.find((s) => s.bucket === 'compute')?.selected_cores || 0);
  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');
  const warmRate = calcRate(warm?.required_storage_tb || 0, warm?.selected_storage_tb || 0);
  const hotRate = calcRate(hot?.required_storage_tb || 0, hot?.selected_storage_tb || 0);
  return { computeRate, warmRate, hotRate };
}

function calcRate(required: number, selected: number): number {
  if (!required || required <= 0) {
    return 100;
  }
  return Math.min(999.9, (selected / required) * 100);
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatMoney(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatPercent(v?: number) {
  return `${(Number(v || 0) * 100).toFixed(2)}%`;
}

function formatSignedPercent(v?: number) {
  const n = Number(v || 0) * 100;
  const prefix = n > 0 ? '+' : '';
  return `${prefix}${n.toFixed(2)}%`;
}

function normalizeConfigType(v?: string) {
  return String(v || '').trim().toLowerCase();
}

function normalizeCountry(v?: string) {
  const raw = String(v || '').trim();
  if (!raw) return '';
  const n = raw.toLowerCase().replace(/[\s_-]/g, '');

  if (['国内', '中国', 'cn', 'china', 'mainlandchina'].includes(n)) return '国内';
  if (['印度', 'india', 'in'].includes(n)) return '印度';

  return raw;
}

function normalizeSceneCategory(v?: string) {
  const raw = String(v || '').trim();
  if (!raw) return 'warm_storage';
  const n = raw.toLowerCase().replace(/[\s_-]/g, '');

  if (['compute', '计算'].includes(n)) return 'compute';
  if (['gpu'].includes(n)) return 'gpu';
  if (['warmstorage', '温存储', '温储', '温', '问存储', 'wenstorage'].includes(n)) return 'warm_storage';
  if (['hotstorage', '热存储', '热储', '热'].includes(n)) return 'hot_storage';

  return raw;
}

function sceneCategoryLabel(v?: string) {
  const key = normalizeSceneCategory(v);
  return SCENE_OPTIONS.find((x) => x.key === key)?.label || v || '-';
}
