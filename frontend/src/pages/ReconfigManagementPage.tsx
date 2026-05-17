import { useMemo, useState } from 'react';
import { Alert, Button, Card, Input, Progress, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { calculateValueScorePerformance, getReconfigPlanProgress, getReconfigPlanResult, listMetaModels, listMetaRecords, startReconfigPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { MetaModel, MetaRecord } from '../types';

const { Text } = Typography;

interface TargetConfig {
  mode: 'existing' | 'maximize';
  configType: string;
  perfBaseline: number;
  memoryDatarateBaseline: number;
  memoryCapacityBaseline: number;
  storageCapacityBaseline: number;
  memoryCpuRatio: number;
}

interface ScopeConfig {
  psaInput: string;
  configTypes: string[];
  snInput: string;
}

interface CandidateRow {
  sn: string;
  configType: string;
  rack: string;
  datacenter: string;
  memoryDatarate: number;
  perfScore: number;
  memoryCapacityGb: number;
  storageCapacityTb: number;
  memoryGapGb: number;
  storageGapTb: number;
  status: '候选' | '内存带宽不足' | '性能不足';
}

interface ActionRow {
  targetSn: string;
  gapType: string;
  gapQty: string;
  source: string;
  partDetails: string;
  crossIdc: '是' | '否';
  action: '拆配' | '改配' | '调拨';
  ruleHit: string;
}

function pick(obj: any, keys: string[], fallback: any = '') {
  for (const k of keys) {
    if (obj?.[k] !== undefined && obj?.[k] !== null && obj?.[k] !== '') return obj[k];
  }
  return fallback;
}

function toNum(v: any, fallback = 0) {
  const n = Number(v);
  return Number.isFinite(n) ? n : fallback;
}

function parseSNList(raw: string): string[] {
  return raw
    .split(/\s+/)
    .map((x) => x.trim())
    .filter(Boolean)
    .slice(0, 1000);
}

export default function ReconfigManagementPage() {
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState<MetaModel[]>([]);
  const [serverRows, setServerRows] = useState<MetaRecord[]>([]);
  const [memoryRows, setMemoryRows] = useState<MetaRecord[]>([]);
  const [configRows, setConfigRows] = useState<MetaRecord[]>([]);
  const [perfByConfig, setPerfByConfig] = useState<Record<string, number>>({});

  const [target, setTarget] = useState<TargetConfig>({
    mode: 'existing',
    configType: '',
    perfBaseline: 0,
    memoryDatarateBaseline: 0,
    memoryCapacityBaseline: 0,
    storageCapacityBaseline: 0,
    memoryCpuRatio: 6
  });
  const [scope, setScope] = useState<ScopeConfig>({
    psaInput: '/server-decommission/cn-decommission/reuse',
    configTypes: [],
    snInput: ''
  });

  const [candidates, setCandidates] = useState<CandidateRow[]>([]);
  const [actions, setActions] = useState<ActionRow[]>([]);
  const [runningJobId, setRunningJobId] = useState<string>('');
  const [progress, setProgress] = useState<any>(null);
  const [planWarnings, setPlanWarnings] = useState<string[]>([]);

  const filteredServers = useMemo(() => {
    const snSet = new Set(parseSNList(scope.snInput));
    return serverRows.filter((r) => {
      const d = r.data || {};
      const psa = String(pick(d, ['psa', 'PSA'], '')).trim();
      const belong = String(pick(d, ['belong', '归属'], '')).trim();
      const configType = String(pick(d, ['config_type', '配置类型'], '')).trim();
      const sn = String(pick(d, ['sn', 'SN'], '')).trim();

      if (belong && belong !== 'ai_data_engineering') return false;
      const psaList = scope.psaInput.split(',').map((x) => x.trim()).filter(Boolean);
      if (psaList.length && !psaList.some((p) => psa.includes(p))) return false;
      if (scope.configTypes.length && !scope.configTypes.includes(configType)) return false;
      if (snSet.size && !snSet.has(sn)) return false;
      return true;
    });
  }, [serverRows, scope]);

  const configTypeOptions = useMemo(() => {
    const s = new Set<string>();
    for (const r of filteredServers) {
      const t = String(pick(r.data, ['config_type', '配置类型'], '')).trim();
      if (t) s.add(t);
    }
    return Array.from(s).sort();
  }, [filteredServers]);

  async function loadAll() {
    setLoading(true);
    try {
      const modelResp = ensureApiOk(await listMetaModels());
      const allModels = modelResp.data.list || [];
      setModels(allModels);

      const modelMap: Record<string, MetaModel | undefined> = {
        server: allModels.find((m) => m.model_code === 'server'),
        rack: allModels.find((m) => m.model_code === 'rack'),
        memory: allModels.find((m) => m.model_code === 'memory'),
        disk: allModels.find((m) => m.model_code === 'disk'),
        config_type: allModels.find((m) => m.model_code === 'config_type')
      };

      const missing = Object.entries(modelMap).filter(([, v]) => !v).map(([k]) => k);
      if (missing.length) {
        message.warning(`缺少模型：${missing.join(', ')}，部分能力不可用`);
      }

      const [server, memory, configType, perf] = await Promise.all([
        modelMap.server ? listMetaRecords(modelMap.server.id) : Promise.resolve({ code: 0, message: 'ok', data: { model: null, fields: [], list: [], total: 0 } } as any),
        modelMap.memory ? listMetaRecords(modelMap.memory.id) : Promise.resolve({ code: 0, message: 'ok', data: { model: null, fields: [], list: [], total: 0 } } as any),
        modelMap.config_type ? listMetaRecords(modelMap.config_type.id) : Promise.resolve({ code: 0, message: 'ok', data: { model: null, fields: [], list: [], total: 0 } } as any),
        calculateValueScorePerformance()
      ]);

      setServerRows((ensureApiOk(server) as any).data.list || []);
      setMemoryRows((ensureApiOk(memory) as any).data.list || []);
      setConfigRows((ensureApiOk(configType) as any).data.list || []);

      const perfResult = (ensureApiOk(perf) as any).data;
      const perfMap: Record<string, number> = {};
      for (const it of perfResult.items || []) {
        perfMap[String(it.config_type || '').trim()] = Number(it.performance_score || 0);
      }
      setPerfByConfig(perfMap);
      message.success('改配管理基础数据已加载');
    } catch (e) {
      message.error(parseApiError(e, '加载基础数据失败'));
    } finally {
      setLoading(false);
    }
  }

  function fillFromConfigType(configType: string) {
    setTarget((prev) => ({ ...prev, configType }));
    if (!configType) return;

    const server = filteredServers.find((r) => String(pick(r.data, ['config_type', '配置类型'], '')).trim() === configType);
    const serverSN = server ? String(pick(server.data, ['sn', 'SN'], '')).trim() : '';

    const firstMem = memoryRows.find((m) => String(pick(m.data, ['sn_server', '服务器SN'], '')).trim() === serverSN);
    const memoryDatarate = toNum(pick(firstMem?.data, ['datarate', '数据传输率(TM/s)', '数据传输率(MT/s)'], 0));

    const pack = configRows.find((r) => String(pick(r.data, ['config_type', '配置类型'], '')).trim() === configType);
    const memoryGb = toNum(pick(pack?.data, ['capacity_memory_gb', '内存容量(GB)'], 0));
    const storageTb = toNum(pick(pack?.data, ['capacity_storage_tb', '存储容量(TB)'], 0));
    const perf = toNum(perfByConfig[configType], 0);

    setTarget((prev) => ({
      ...prev,
      configType,
      memoryDatarateBaseline: memoryDatarate,
      memoryCapacityBaseline: memoryGb,
      storageCapacityBaseline: storageTb,
      perfBaseline: perf
    }));
  }

  async function computePlan() {
    try {
      const payload = {
        target: {
          mode: target.mode,
          config_type: target.configType,
          perf_baseline: target.perfBaseline,
          memory_datarate_baseline: target.memoryDatarateBaseline,
          memory_capacity_baseline: target.memoryCapacityBaseline,
          storage_capacity_baseline: target.storageCapacityBaseline,
          memory_cpu_ratio: target.memoryCpuRatio
        },
        scope: {
          psa_list: scope.psaInput.split(',').map((x) => x.trim()).filter(Boolean),
          config_types: scope.configTypes,
          sn_input: scope.snInput
        }
      };
      const startResp = ensureApiOk(await startReconfigPlan(payload));
      const jobId = String(startResp.data?.job_id || '');
      if (!jobId) throw new Error('任务启动失败：缺少 job_id');
      setRunningJobId(jobId);
      setProgress({ stage: 'queued', percent: 0, done_packages: 0, total_packages: 0, done_servers: 0, total_servers: 0, done_cores: 0, total_cores: 0 });
      message.info('改配任务已启动，可切换页面，后台会继续执行');

      // 轮询任务进度
      // eslint-disable-next-line no-constant-condition
      while (true) {
        const progResp = ensureApiOk(await getReconfigPlanProgress(jobId));
        const pd = progResp.data || {};
        setProgress(pd.progress || null);
        const status = String(pd.status || 'running');
        if (status === 'running') {
          await new Promise((resolve) => window.setTimeout(resolve, 1000));
          continue;
        }
        if (status === 'failed') {
          throw new Error(String(pd.error || '计算失败'));
        }

        const resultResp = ensureApiOk(await getReconfigPlanResult(jobId));
        const data = resultResp.data?.result || {};
        setTarget((prev) => ({
          ...prev,
          perfBaseline: Number(data.target_resolved?.perf_baseline || prev.perfBaseline),
          memoryDatarateBaseline: Number(data.target_resolved?.memory_datarate_baseline || prev.memoryDatarateBaseline),
          memoryCapacityBaseline: Number(data.target_resolved?.memory_capacity_baseline || prev.memoryCapacityBaseline),
          storageCapacityBaseline: Number(data.target_resolved?.storage_capacity_baseline || prev.storageCapacityBaseline)
        }));
        setCandidates((data.candidates || []).map((x: any) => ({
          sn: x.sn,
          configType: x.config_type,
          rack: x.rack,
          datacenter: x.datacenter,
          memoryDatarate: Number(x.memory_datarate || 0),
          perfScore: Number(x.perf_score || 0),
          memoryCapacityGb: Number(x.memory_capacity_gb || 0),
          storageCapacityTb: Number(x.storage_capacity_tb || 0),
          memoryGapGb: Number(x.memory_gap_gb || 0),
          storageGapTb: Number(x.storage_gap_tb || 0),
          status: x.status
        })));
        setActions((data.actions || []).map((x: any) => ({
          targetSn: x.target_sn,
          gapType: x.gap_type,
          gapQty: x.gap_qty,
          source: x.source,
          partDetails: x.part_details,
          crossIdc: x.cross_idc,
          action: x.action,
          ruleHit: x.rule_hit
        })));
        setPlanWarnings(Array.isArray(data.summary?.warnings) ? data.summary.warnings : []);
        message.success(`计算完成：候选 ${Number(data.summary?.candidate_count || 0)} 台，执行项 ${Number(data.summary?.action_count || 0)} 条`);
        break;
      }
    } catch (e) {
      setRunningJobId('');
      setPlanWarnings([]);
      message.error(parseApiError(e, '计算改配方案失败'));
    }
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Card title="改配管理（首版）" extra={<Button loading={loading} onClick={loadAll}>加载基础数据</Button>}>
        <Alert
          type="info"
          showIcon
          message="依据《02030202-改配管理》文档实现首版：目标配置、范围筛选、候选识别、执行清单。"
          description="当前数据来源：元数据模型（server/rack/memory/disk/config_type）+ 价值分性能跑分。"
        />
      </Card>

      <Card title="目标配置（6.1）">
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Space wrap>
            <Text>套餐模式</Text>
            <Select
              value={target.mode}
              style={{ width: 160 }}
              options={[{ value: 'existing', label: '指定套餐' }, { value: 'maximize', label: '最大化利用' }]}
              onChange={(v) => setTarget({ ...target, mode: v as TargetConfig['mode'] })}
            />
            <Text>配置类型</Text>
            <Select
              value={target.configType}
              style={{ width: 240 }}
              placeholder="请选择配置类型"
              options={configTypeOptions.map((x) => ({ value: x, label: x }))}
              onChange={fillFromConfigType}
              showSearch
              optionFilterProp="label"
              disabled={target.mode === 'maximize'}
            />
          </Space>
          <Space wrap>
            <Text>性能基线(E/s)</Text>
            <Input
              style={{ width: 140 }}
              value={String(target.perfBaseline || '')}
              disabled={target.mode === 'maximize'}
              onChange={(e) => setTarget({ ...target, perfBaseline: toNum(e.target.value) })}
            />
            <Text>内存速率基线(MT/s)</Text>
            <Input
              style={{ width: 140 }}
              value={String(target.memoryDatarateBaseline || '')}
              disabled={target.mode === 'maximize'}
              onChange={(e) => setTarget({ ...target, memoryDatarateBaseline: toNum(e.target.value) })}
            />
            <Text>内存容量基线(GB)</Text>
            <Input
              style={{ width: 140 }}
              value={String(target.memoryCapacityBaseline || '')}
              disabled={target.mode === 'maximize'}
              onChange={(e) => setTarget({ ...target, memoryCapacityBaseline: toNum(e.target.value) })}
            />
            <Text>存储容量基线(TB)</Text>
            <Input
              style={{ width: 140 }}
              value={String(target.storageCapacityBaseline || '')}
              disabled={target.mode === 'maximize'}
              onChange={(e) => setTarget({ ...target, storageCapacityBaseline: toNum(e.target.value) })}
            />
            <Text>内存CPU比</Text>
            <Input
              style={{ width: 120 }}
              value={String(target.memoryCpuRatio || '')}
              onChange={(e) => setTarget({ ...target, memoryCpuRatio: toNum(e.target.value, 6) })}
            />
          </Space>
        </Space>
      </Card>

      <Card title="范围配置（6.2）" extra={<Button type="primary" loading={!!runningJobId} onClick={computePlan}>计算改配方案</Button>}>
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Space wrap>
            <Text>PSA（单值；多个用英文逗号分隔）</Text>
            <Input
              style={{ width: 520 }}
              value={scope.psaInput}
              onChange={(e) => setScope({ ...scope, psaInput: e.target.value })}
              placeholder="示例：/server-decommission/cn-decommission/reuse 或 /a,/b,/c"
            />
          </Space>
          <Space wrap>
            <Text>配置类型（默认全选）</Text>
            <Select
              mode="multiple"
              style={{ width: 520 }}
              placeholder="不选=全选"
              value={scope.configTypes}
              onChange={(v) => setScope({ ...scope, configTypes: v as string[] })}
              options={configTypeOptions.map((x) => ({ value: x, label: x }))}
              showSearch
              optionFilterProp="label"
            />
          </Space>
          <Space wrap>
            <Text>SN筛选（空格分隔，最多1000）</Text>
            <Input.TextArea
              rows={2}
              style={{ width: 720 }}
              value={scope.snInput}
              onChange={(e) => setScope({ ...scope, snInput: e.target.value })}
              placeholder="如：SN001 SN002 SN003"
            />
          </Space>
          <Text type="secondary">范围命中：{filteredServers.length} 台</Text>
          {runningJobId && progress ? (
            <Card size="small" style={{ marginTop: 8, width: 760 }}>
              <Space direction="vertical" style={{ width: '100%' }} size={6}>
                <Text strong>任务执行中：{runningJobId}</Text>
                <Text type="secondary">阶段：{progress.stage || '-'} {progress.message ? `｜${progress.message}` : ''}</Text>
                <Progress percent={Number(progress.percent || 0)} status="active" />
                <Space wrap>
                  <Tag color="blue">套餐进度：{Number(progress.done_packages || 0)}/{Number(progress.total_packages || 0)}</Tag>
                  <Tag color="green">主机进度：{Number(progress.done_servers || 0)}/{Number(progress.total_servers || 0)}</Tag>
                  <Tag color="purple">逻辑核进度：{Number(progress.done_cores || 0)}/{Number(progress.total_cores || 0)}</Tag>
                </Space>
              </Space>
            </Card>
          ) : null}
        </Space>
      </Card>

      {planWarnings.length > 0 ? (
        <Alert
          type="warning"
          showIcon
          message="数据告警"
          description={planWarnings.join('；')}
        />
      ) : null}

      <Tabs
        items={[
          {
            key: 'candidate',
            label: `改配候选清单（${candidates.length}）`,
            children: (
              <Table
                rowKey={(r) => `${r.sn}-${r.status}`}
                size="small"
                pagination={{ pageSize: 20 }}
                dataSource={candidates}
                columns={[
                  { title: 'SN', dataIndex: 'sn', width: 180, fixed: 'left' as const },
                  { title: '配置类型', dataIndex: 'configType', width: 160 },
                  { title: '机房', dataIndex: 'datacenter', width: 140 },
                  { title: '内存速率', dataIndex: 'memoryDatarate', width: 110 },
                  { title: '性能跑分', dataIndex: 'perfScore', width: 110 },
                  { title: '当前内存(GB)', dataIndex: 'memoryCapacityGb', width: 120 },
                  { title: '当前存储(TB)', dataIndex: 'storageCapacityTb', width: 120 },
                  { title: '内存缺口(GB)', dataIndex: 'memoryGapGb', width: 120 },
                  { title: '存储缺口(TB)', dataIndex: 'storageGapTb', width: 120 },
                  {
                    title: '状态', dataIndex: 'status', width: 120, render: (v: CandidateRow['status']) => {
                      if (v === '候选') return <Tag color="green">候选</Tag>;
                      if (v === '内存带宽不足') return <Tag color="orange">内存带宽不足</Tag>;
                      return <Tag color="red">性能不足</Tag>;
                    }
                  }
                ]}
                scroll={{ x: 1500 }}
              />
            )
          },
          {
            key: 'actions',
            label: `执行清单（${actions.length}）`,
            children: (
              <Table
                rowKey={(r) => `${r.targetSn}-${r.gapType}-${r.partDetails}`}
                size="small"
                pagination={{ pageSize: 20 }}
                dataSource={actions}
                columns={[
                  { title: '目标SN', dataIndex: 'targetSn', width: 180, fixed: 'left' as const },
                  { title: '缺口类型', dataIndex: 'gapType', width: 120 },
                  { title: '缺口数量', dataIndex: 'gapQty', width: 100 },
                  { title: '推荐来源', dataIndex: 'source', width: 180 },
                  { title: '配件明细', dataIndex: 'partDetails', width: 260 },
                  { title: '跨机房搬迁', dataIndex: 'crossIdc', width: 120 },
                  { title: '执行动作', dataIndex: 'action', width: 100 },
                  { title: '规则命中说明', dataIndex: 'ruleHit', width: 320 }
                ]}
                scroll={{ x: 1500 }}
              />
            )
          },
          {
            key: 'meta',
            label: '依赖模型状态',
            children: (
              <Table
                rowKey={(r) => r.id}
                size="small"
                pagination={false}
                dataSource={models.filter((m) => ['server', 'rack', 'memory', 'disk', 'config_type'].includes(m.model_code))}
                columns={[
                  { title: '模型编码', dataIndex: 'model_code', width: 140 },
                  { title: '模型名称', dataIndex: 'model_name', width: 160 },
                  { title: '状态', dataIndex: 'status', width: 100 },
                  { title: '版本', dataIndex: 'current_version', width: 90 },
                  { title: '更新时间', dataIndex: 'updated_at' }
                ]}
              />
            )
          }
        ]}
      />
    </Space>
  );
}
