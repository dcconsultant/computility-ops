import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { Alert, Button, Card, Input, message, Space, Table, Tabs, Typography, Upload } from 'antd';
import type { UploadProps } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import {
  exportServerPackageAnomalies,
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

type DataKey = 'servers' | 'packages' | 'assets';

const titles: Record<DataKey, string> = {
  servers: '服务器管理',
  packages: '主机套餐配置',
  assets: '资产分析'
};

export default function ImportPage() {
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [uploading, setUploading] = useState<DataKey | null>(null);

  const [servers, setServers] = useState<ServerItem[]>([]);
  const [packages, setPackages] = useState<HostPackageConfig[]>([]);
  const [serverKeyword, setServerKeyword] = useState('');
  const [packageKeyword, setPackageKeyword] = useState('');

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
      x.server_value_score,
      x.arch_standardized_factor
    ].some((v) => String(v ?? '').toLowerCase().includes(q)));
  }, [packages, packageKeyword]);

  const assetAnalysis = useMemo(() => buildAssetAnalysis(servers), [servers]);

  function makeUploadProps(kind: Exclude<DataKey, 'assets'>): UploadProps {
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

  const tableCard = (title: string, kind: Exclude<DataKey, 'assets'>, table: ReactNode, helper: string, extra?: ReactNode) => (
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
                <Input
                  allowClear
                  placeholder="搜索服务器（SN/型号/PSA/环境/配置类型等）"
                  value={serverKeyword}
                  onChange={(e) => setServerKeyword(e.target.value)}
                />
                <Table
                  rowKey="sn"
                  dataSource={filteredServers}
                  pagination={withTotalPagination(10)}
                  columns={[
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
                  ]}
                />
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
                <Input
                  allowClear
                  placeholder="搜索套餐（配置类型/场景/核数/卡数/存储等）"
                  value={packageKeyword}
                  onChange={(e) => setPackageKeyword(e.target.value)}
                />
                <Table
                  rowKey="config_type"
                  dataSource={filteredPackages}
                  pagination={withTotalPagination(10)}
                  columns={[
                    { title: '配置类型', dataIndex: 'config_type' },
                    { title: '场景大类', dataIndex: 'scene_category' },
                    { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores', render: (v: number) => formatInt(v) },
                    { title: 'GPU卡数', dataIndex: 'gpu_card_count', render: (v: number) => formatInt(v) },
                    { title: '数据盘类型', dataIndex: 'data_disk_type' },
                    { title: '数据盘数量', dataIndex: 'data_disk_count', render: (v: number) => formatInt(v) },
                    { title: '存储容量(TB)', dataIndex: 'storage_capacity_tb', render: (v: number) => formatFloat(v) },
                    { title: '服务器价值分', dataIndex: 'server_value_score', render: (v: number) => formatFloat(v) },
                    { title: '架构标准化系数', dataIndex: 'arch_standardized_factor', render: (v: number) => formatFloat(v) }
                  ]}
                />
              </Space>,
              '服务器管理表通过配置类型关联此表；需维护服务器价值分（PSA非数字时基准）与GPU卡数（GPU汇总统计依赖）。'
            )
          },
          {
            key: 'assets',
            label: '资产分析',
            children: (
              <Space direction="vertical" size="large" style={{ width: '100%' }}>
                <Card title="国内/印度保内保外概览">
                  <Table
                    rowKey={(r) => `${r.region}-${r.snapshotDate}`}
                    dataSource={assetAnalysis.snapshotRows}
                    pagination={false}
                    size="small"
                    columns={[
                      { title: '地区', dataIndex: 'region' },
                      { title: '统计时点', dataIndex: 'snapshotLabel' },
                      { title: '日期', dataIndex: 'snapshotDate' },
                      { title: '保内', dataIndex: 'inWarranty', render: (v: number) => formatInt(v) },
                      { title: '保外', dataIndex: 'outWarranty', render: (v: number) => formatInt(v) },
                      { title: '总数量', dataIndex: 'total', render: (v: number) => formatInt(v) },
                      { title: '累计过保占比', dataIndex: 'outWarrantyRatio', render: (v: number) => `${v.toFixed(2)}%` }
                    ]}
                  />
                  <Text type="secondary">口径：IDC 以 "IN" 开头判定为印度，其余归为国内；日期为空/异常按已过保处理。</Text>
                </Card>

                <Card title="国内服务器过保组合趋势图">
                  <AssetTrendChart points={assetAnalysis.trends.domestic} total={assetAnalysis.totals.domestic} regionLabel="国内" />
                </Card>

                <Card title="印度服务器过保组合趋势图">
                  <AssetTrendChart points={assetAnalysis.trends.india} total={assetAnalysis.totals.india} regionLabel="印度" />
                </Card>
              </Space>
            )
          }
        ]}
      />
    </Space>
  );
}

type RegionKey = 'domestic' | 'india';

interface AssetSnapshotRow {
  region: '国内' | '印度';
  snapshotLabel: string;
  snapshotDate: string;
  inWarranty: number;
  outWarranty: number;
  total: number;
  outWarrantyRatio: number;
}

interface AssetTrendPoint {
  year: number;
  outCount: number;
  cumulativeOutCount: number;
  cumulativeOutRatio: number;
}

function buildAssetAnalysis(servers: ServerItem[]) {
  const now = new Date();
  const nextYear0630 = new Date(now.getFullYear() + 1, 5, 30);
  const snapshots = [
    { label: '当前时间', date: now },
    { label: '次年6月30日', date: nextYear0630 }
  ];

  const totals: Record<RegionKey, number> = { domestic: 0, india: 0 };
  const snapshotRows: AssetSnapshotRow[] = [];
  const trends: Record<RegionKey, AssetTrendPoint[]> = {
    domestic: [],
    india: []
  };

  (['domestic', 'india'] as RegionKey[]).forEach((region) => {
    const list = servers.filter((s) => resolveRegion(s.idc) === region);
    totals[region] = list.length;

    for (const snap of snapshots) {
      let outWarranty = 0;
      for (const item of list) {
        const end = parseYMD(item.warranty_end_date);
        if (!end || end.getTime() < snap.date.getTime()) {
          outWarranty += 1;
        }
      }
      const total = list.length;
      snapshotRows.push({
        region: region === 'domestic' ? '国内' : '印度',
        snapshotLabel: snap.label,
        snapshotDate: formatDate(snap.date),
        inWarranty: Math.max(0, total - outWarranty),
        outWarranty,
        total,
        outWarrantyRatio: total > 0 ? (outWarranty * 100) / total : 0
      });
    }

    const yearMap = new Map<number, number>();
    for (const item of list) {
      const end = parseYMD(item.warranty_end_date);
      if (!end) continue;
      const y = end.getFullYear();
      yearMap.set(y, (yearMap.get(y) || 0) + 1);
    }

    let cumulative = 0;
    trends[region] = [...yearMap.entries()]
      .sort((a, b) => a[0] - b[0])
      .map(([year, outCount]) => {
        cumulative += outCount;
        return {
          year,
          outCount,
          cumulativeOutCount: cumulative,
          cumulativeOutRatio: list.length > 0 ? (cumulative * 100) / list.length : 0
        };
      });
  });

  return { snapshotRows, trends, totals };
}

function resolveRegion(idc?: string): RegionKey {
  const norm = (idc || '').trim().toUpperCase();
  return norm.startsWith('IN') ? 'india' : 'domestic';
}

function parseYMD(v?: string) {
  if (!v) return null;
  const m = /^\s*(\d{4})-(\d{2})-(\d{2})\s*$/.exec(v);
  if (!m) return null;
  const y = Number(m[1]);
  const mon = Number(m[2]);
  const d = Number(m[3]);
  if (!Number.isFinite(y) || mon < 1 || mon > 12 || d < 1 || d > 31) return null;
  return new Date(y, mon - 1, d);
}

function formatDate(d: Date) {
  const y = d.getFullYear();
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function AssetTrendChart({ points, total, regionLabel }: { points: AssetTrendPoint[]; total: number; regionLabel: string }) {
  if (!points.length) {
    return <Text type="secondary">暂无可用于绘图的过保日期数据。</Text>;
  }

  const width = 880;
  const height = 320;
  const m = { left: 52, right: 54, top: 20, bottom: 54 };
  const innerW = width - m.left - m.right;
  const innerH = height - m.top - m.bottom;
  const maxCount = Math.max(1, ...points.map((p) => p.outCount));

  const x = (idx: number) => m.left + ((idx + 0.5) * innerW) / points.length;
  const barW = Math.max(14, Math.min(42, innerW / Math.max(1, points.length) * 0.55));
  const yCount = (v: number) => m.top + innerH - (v / maxCount) * innerH;
  const yRatio = (v: number) => m.top + innerH - (v / 100) * innerH;

  const linePath = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i)},${yRatio(p.cumulativeOutRatio)}`)
    .join(' ');

  const labelMinGap = 10;
  const labelTop = m.top + 10;
  const labelBottom = m.top + innerH - 2;
  const clampLabelY = (v: number) => Math.max(labelTop, Math.min(labelBottom, v));
  const labelPositions = points.map((p) => {
    let barY = clampLabelY(yCount(p.outCount) - 6);
    let lineY = clampLabelY(yRatio(p.cumulativeOutRatio) - 8);

    if (Math.abs(barY - lineY) < labelMinGap) {
      const linePreferAbove = lineY <= barY;
      if (linePreferAbove) {
        lineY = clampLabelY(barY - labelMinGap);
        if (Math.abs(barY - lineY) < labelMinGap) {
          barY = clampLabelY(lineY + labelMinGap);
        }
      } else {
        lineY = clampLabelY(barY + labelMinGap);
        if (Math.abs(barY - lineY) < labelMinGap) {
          barY = clampLabelY(lineY - labelMinGap);
        }
      }
    }

    return { barY, lineY };
  });

  return (
    <Space direction="vertical" size={12} style={{ width: '100%' }}>
      <svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', background: '#fff', borderRadius: 8 }}>
        {[0, 25, 50, 75, 100].map((r) => (
          <g key={r}>
            <line x1={m.left} x2={width - m.right} y1={yRatio(r)} y2={yRatio(r)} stroke="#f0f0f0" strokeWidth="1" />
            <text x={width - m.right + 6} y={yRatio(r) + 4} fontSize="10" fill="#888">{r}%</text>
          </g>
        ))}

        {points.map((p, i) => (
          <g key={p.year}>
            <rect
              x={x(i) - barW / 2}
              y={yCount(p.outCount)}
              width={barW}
              height={Math.max(0, m.top + innerH - yCount(p.outCount))}
              fill="#91caff"
              rx="4"
            >
              <title>{`${p.year}年过保数量：${p.outCount}`}</title>
            </rect>
            <text x={x(i)} y={labelPositions[i].barY} textAnchor="middle" fontSize="10" fill="#3b6ea8">{p.outCount}</text>
            <text x={x(i)} y={height - 22} textAnchor="middle" fontSize="11" fill="#666">{p.year}</text>
          </g>
        ))}

        <path d={linePath} fill="none" stroke="#ff4d4f" strokeWidth="2.5" />
        {points.map((p, i) => (
          <g key={`dot-${p.year}`}>
            <circle cx={x(i)} cy={yRatio(p.cumulativeOutRatio)} r="4" fill="#ff4d4f" />
            <text x={x(i)} y={labelPositions[i].lineY} textAnchor="middle" fontSize="10" fill="#c62828">
              {`${p.cumulativeOutRatio.toFixed(2)}%`}
            </text>
            <title>{`${p.year}年累计过保占比：${p.cumulativeOutRatio.toFixed(2)}%`}</title>
          </g>
        ))}

        <text x={m.left - 40} y={m.top - 2} fontSize="10" fill="#888">过保数量</text>
        <text x={width - m.right + 6} y={m.top - 2} fontSize="10" fill="#888">累计占比</text>
      </svg>

      <Table
        size="small"
        pagination={false}
        rowKey={(r) => `${regionLabel}-${r.year}`}
        dataSource={points}
        columns={[
          { title: '年份', dataIndex: 'year' },
          { title: '当年过保数量', dataIndex: 'outCount', render: (v: number) => formatInt(v) },
          { title: '累计过保数量', dataIndex: 'cumulativeOutCount', render: (v: number) => formatInt(v) },
          { title: '累计过保占比', dataIndex: 'cumulativeOutRatio', render: (v: number) => `${v.toFixed(2)}%` }
        ]}
      />
      <Text type="secondary">{regionLabel}样本总量：{formatInt(total)} 台</Text>
    </Space>
  );
}

function withTotalPagination(pageSize: number) {
  return {
    pageSize,
    showTotal: (total: number) => `共${total}条，${Math.ceil(total / pageSize)}页`
  };
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatMaybeNumber(v?: string) {
  const n = Number((v || '').trim());
  if (Number.isNaN(n)) return v || '-';
  return n.toLocaleString('en-US', { maximumFractionDigits: 2 });
}
