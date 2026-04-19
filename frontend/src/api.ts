import axios from 'axios';
import type {
  ApiResp,
  HostPackageConfig,
  ImportResult,
  ListData,
  ModelFailureRate,
  FaultAnalysisResult,
  FaultYearAnalysisRow,
  FailureRateSummary,
  FailureOverviewCard,
  FailureAgeTrendPoint,
  FailureFeatureFact,
  PackageFailureRate,
  PackageModelFailureRate,
  RenewalPlan,
  ServerItem,
  SpecialRule,
  StorageBucket,
  StorageTopServerRate,
  ImportErrorInsight
} from './types';

const http = axios.create({ baseURL: '/api/v1' });

async function uploadImport(url: string, file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<ImportResult>>(url, form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export async function importServers(file: File) {
  return uploadImport('/servers/import', file);
}
export async function listServers() {
  const { data } = await http.get<ApiResp<ListData<ServerItem>>>('/servers');
  return data;
}

export function exportServerPackageAnomalies(format: 'xlsx' | 'csv' = 'xlsx') {
  window.open(`/api/v1/servers/package-anomalies/export?format=${format}`, '_blank');
}

export async function importHostPackages(file: File) {
  return uploadImport('/host-packages/import', file);
}
export async function listHostPackages() {
  const { data } = await http.get<ApiResp<ListData<HostPackageConfig>>>('/host-packages');
  return data;
}

export async function importSpecialRules(file: File) {
  return uploadImport('/special-rules/import', file);
}
export async function listSpecialRules() {
  const { data } = await http.get<ApiResp<ListData<SpecialRule>>>('/special-rules');
  return data;
}

export async function importModelFailureRates(file: File) {
  return uploadImport('/failure-rates/model/import', file);
}
export async function listModelFailureRates() {
  const { data } = await http.get<ApiResp<ListData<ModelFailureRate>>>('/failure-rates/model');
  return data;
}

export async function importPackageFailureRates(file: File) {
  return uploadImport('/failure-rates/package/import', file);
}
export async function listPackageFailureRates() {
  const { data } = await http.get<ApiResp<ListData<PackageFailureRate>>>('/failure-rates/package');
  return data;
}

export async function importPackageModelFailureRates(file: File) {
  return uploadImport('/failure-rates/package-model/import', file);
}
export async function listPackageModelFailureRates() {
  const { data } = await http.get<ApiResp<ListData<PackageModelFailureRate>>>('/failure-rates/package-model');
  return data;
}

export async function analyzeFaultRates(file: File, opts?: { excludeOverWarranty?: boolean }) {
  const { data } = await http.post<ApiResp<FaultAnalysisResult>>('/failure-rates/analyze/import', (() => {
    const form = new FormData();
    form.append('file', file);
    if (opts?.excludeOverWarranty) {
      form.append('exclude_over_warranty', 'true');
    }
    return form;
  })(), {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export async function listOverallFailureRates() {
  const { data } = await http.get<ApiResp<ListData<FailureRateSummary>>>('/failure-rates/overall');
  return data;
}

export async function listFailureOverviewCards() {
  const { data } = await http.get<ApiResp<ListData<FailureOverviewCard>>>('/failure-rates/overview-cards');
  return data;
}

export async function listFailureAgeTrendPoints() {
  const { data } = await http.get<ApiResp<ListData<FailureAgeTrendPoint>>>('/failure-rates/age-trend');
  return data;
}

export async function listFailureFeatureFacts() {
  const { data } = await http.get<ApiResp<ListData<FailureFeatureFact>>>('/failure-rates/features');
  return data;
}

export async function listStorageTopServerRates(bucket: StorageBucket = 'warm_storage') {
  const { data } = await http.get<ApiResp<ListData<StorageTopServerRate>>>('/failure-rates/storage-top-servers', {
    params: { bucket }
  });
  return data;
}

export function exportStorageTopServers(bucket: StorageBucket = 'warm_storage', format: 'xlsx' | 'csv' = 'xlsx') {
  window.open(`/api/v1/failure-rates/storage-top-servers/export?bucket=${bucket}&format=${format}`, '_blank');
}

export function exportWarmStorageServers(format: 'xlsx' | 'csv' = 'xlsx') {
  exportStorageTopServers('warm_storage', format);
}

export async function exportYearFaultAnalysis(rows: FaultYearAnalysisRow[], year = new Date().getFullYear()) {
  const { data } = await http.post('/failure-rates/year-fault-analysis/export', { year, rows }, {
    responseType: 'blob'
  });
  return data as Blob;
}

export interface CreatePlanPayload {
  target_date: string;
  excluded_environments: string[];
  excluded_psas: string[];
  target_cores: number;
  warm_target_storage_tb: number;
  hot_target_storage_tb: number;
}

export async function createPlan(payload: CreatePlanPayload) {
  const { data } = await http.post<ApiResp<RenewalPlan>>('/renewals/plan', payload);
  return data;
}

export interface ListPlansParams {
  plan_id?: string;
  target_date_from?: string;
  target_date_to?: string;
  excluded_psa?: string;
  excluded_environment?: string;
}

export async function listPlans(params?: ListPlansParams) {
  const { data } = await http.get<ApiResp<ListData<RenewalPlan>>>('/renewals/plans', { params });
  return data;
}

export async function getPlan(planId: string) {
  const { data } = await http.get<ApiResp<RenewalPlan>>(`/renewals/plans/${planId}`);
  return data;
}

export async function deletePlan(planId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; plan_id: string }>>(`/renewals/plans/${planId}`);
  return data;
}

export function exportPlan(planId: string, format: 'xlsx' | 'csv') {
  window.open(`/api/v1/renewals/plans/${planId}/export?format=${format}`, '_blank');
}

export interface MySQLTestPayload {
  dsn?: string;
  host?: string;
  port?: number;
  user?: string;
  password?: string;
  database?: string;
  params?: string;
}

export interface MySQLTestResult {
  reachable: boolean;
  latency_ms: number;
  message: string;
}

export async function testMySQLConnection(payload: MySQLTestPayload) {
  const { data } = await http.post<ApiResp<MySQLTestResult>>('/system/mysql/test', payload);
  return data;
}

export async function listImportErrors(limit = 20) {
  const { data } = await http.get<ApiResp<ListData<ImportErrorInsight>>>(`/system/import-errors?limit=${limit}`);
  return data;
}
