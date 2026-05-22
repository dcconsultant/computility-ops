import axios from 'axios';
import type {
  ApiResp,
  HostPackageConfig,
  ImportResult,
  ListData,
  CabinetConfig,
  CabinetUtilizationSetting,
  ValueScoreCabinetBaseline,
  ValueScoreCostParams,
  ValueScoreOriginalValue,
  ValueScorePerformanceCalcResult,
  ValueScorePerformanceParam,
  ValueScoreTCOResult,
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
  RenewalPlanSettings,
  RenewalRequirements,
  RenewalUnitPrice,
  ServerItem,
  SpecialRule,
  StorageBucket,
  StorageTopServerRate,
  ImportErrorInsight,
  Contract,
  ContractAttachment,
  Supplier,
  MetaField,
  MetaModel,
  MetaReference,
  MetaModelVersion,
  MetaModelSnapshot,
  MetaEnumOption,
  MetaRecord,
  ResourcePlanningRequest,
  ResourcePlanningResponse,
  ResourcePlanningConfigState
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

export function exportHostPackageTemplate() {
  window.open('/api/v1/host-packages/template/export', '_blank');
}

export async function getCabinetUtilization() {
  const { data } = await http.get<ApiResp<CabinetUtilizationSetting>>('/cabinet-config/utilization');
  return data;
}

export async function updateCabinetUtilization(utilization: number) {
  const { data } = await http.put<ApiResp<CabinetUtilizationSetting>>('/cabinet-config/utilization', { utilization });
  return data;
}

export async function importCabinetConfigs(file: File) {
  return uploadImport('/cabinet-config/import', file);
}

export function exportCabinetTemplate() {
  window.open('/api/v1/cabinet-config/template/export', '_blank');
}

export async function listCabinetConfigs() {
  const { data } = await http.get<ApiResp<ListData<CabinetConfig>>>('/cabinet-config');
  return data;
}

export async function createCabinetConfig(payload: Omit<CabinetConfig, 'id'>) {
  const { data } = await http.post<ApiResp<CabinetConfig>>('/cabinet-config', payload);
  return data;
}

export async function updateCabinetConfig(id: number, payload: Omit<CabinetConfig, 'id'>) {
  const { data } = await http.put<ApiResp<CabinetConfig>>(`/cabinet-config/${id}`, payload);
  return data;
}

export async function deleteCabinetConfig(id: number) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; id: number }>>(`/cabinet-config/${id}`);
  return data;
}

export async function getValueScoreCabinetBaseline() {
  const { data } = await http.get<ApiResp<ValueScoreCabinetBaseline>>('/value-score/cabinet-baseline');
  return data;
}

export async function getValueScoreCostParams() {
  const { data } = await http.get<ApiResp<ValueScoreCostParams>>('/value-score/cost-params');
  return data;
}

export async function updateValueScoreCostParams(payload: ValueScoreCostParams) {
  const { data } = await http.put<ApiResp<ValueScoreCostParams>>('/value-score/cost-params', payload);
  return data;
}

export async function calculateValueScoreTCO(configTypes?: string[]) {
  const { data } = await http.post<ApiResp<ValueScoreTCOResult>>('/value-score/tco/calculate', { config_types: configTypes || [] });
  return data;
}

export function exportValueScoreCostParamsTemplate() {
  window.open('/api/v1/value-score/cost-params/template/export', '_blank');
}

export async function importValueScoreCostParams(file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<ValueScoreCostParams>>('/value-score/cost-params/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export function exportValueScoreOriginalValueTemplate() {
  window.open('/api/v1/value-score/original-values/template/export', '_blank');
}

export function exportValueScorePerformanceParamsTemplate() {
  window.open('/api/v1/value-score/performance-params/template/export', '_blank');
}

export async function listValueScorePerformanceParams() {
  const { data } = await http.get<ApiResp<{ list: ValueScorePerformanceParam[] }>>('/value-score/performance-params');
  return data;
}

export async function previewValueScorePerformanceParams(file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<any>>('/value-score/performance-params/preview', form, { headers: { 'Content-Type': 'multipart/form-data' } });
  return data;
}

export async function importValueScoreUnifiedParams(file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<any>>('/value-score/performance-params/import', form, { headers: { 'Content-Type': 'multipart/form-data' } });
  return data;
}

export async function calculateValueScorePerformance() {
  const { data } = await http.post<ApiResp<ValueScorePerformanceCalcResult>>('/value-score/performance/calculate', {});
  return data;
}

export async function getResourcePlanningConfig() {
  const { data } = await http.get<ApiResp<{ found: boolean; state?: ResourcePlanningConfigState }>>('/resource-planning/config');
  return data;
}

export async function saveResourcePlanningConfig(payload: ResourcePlanningRequest) {
  const { data } = await http.post<ApiResp<{ saved: boolean }>>('/resource-planning/config', payload);
  return data;
}

export async function calculateResourcePlanning(payload: ResourcePlanningRequest) {
  const { data } = await http.post<ApiResp<ResourcePlanningResponse>>('/resource-planning/calculate', payload);
  return data;
}

export async function calculateReconfigPlan(payload: any) {
  const { data } = await http.post<ApiResp<any>>('/reconfig/plan/calculate', payload);
  return data;
}

export async function startReconfigPlan(payload: any) {
  const { data } = await http.post<ApiResp<{ job_id: string; status: string }>>('/reconfig/plan/start', payload);
  return data;
}

export async function getReconfigPlanProgress(jobId: string) {
  const { data } = await http.get<ApiResp<any>>(`/reconfig/plan/progress/${jobId}`);
  return data;
}

export async function getReconfigPlanResult(jobId: string) {
  const { data } = await http.get<ApiResp<any>>(`/reconfig/plan/result/${jobId}`);
  return data;
}

export async function listSavedReconfigPlans() {
  const { data } = await http.get<ApiResp<{ list: any[]; total: number }>>('/reconfig/plans');
  return data;
}

export async function getSavedReconfigPlan(planId: string) {
  const { data } = await http.get<ApiResp<any>>(`/reconfig/plans/${planId}`);
  return data;
}

export function exportReconfigActionsByJob(jobId: string) {
  window.open(`/api/v1/reconfig/plan/result/${jobId}/actions/export`, '_blank');
}

export function exportReconfigActionsByPlan(planId: string) {
  window.open(`/api/v1/reconfig/plans/${planId}/actions/export`, '_blank');
}

export async function importValueScoreOriginalValues(file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<{ imported: number }>>('/value-score/original-values/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export async function listValueScoreOriginalValues() {
  const { data } = await http.get<ApiResp<{ list: ValueScoreOriginalValue[] }>>('/value-score/original-values');
  return data;
}

export function exportValueScoreTCO() {
  window.open('/api/v1/value-score/tco/export', '_blank');
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
  target_cores?: number;
  warm_target_storage_tb?: number;
  hot_target_storage_tb?: number;
  requirements: RenewalRequirements;
  domestic_budget: number;
  india_budget: number;
}

export async function createPlan(payload: CreatePlanPayload) {
  const { data } = await http.post<ApiResp<RenewalPlan>>('/renewals/plan', payload);
  return data;
}

export async function getRenewalSettings() {
  const { data } = await http.get<ApiResp<RenewalPlanSettings>>('/renewals/settings');
  return data;
}

export async function updateRenewalSettings(payload: RenewalPlanSettings) {
  const { data } = await http.put<ApiResp<RenewalPlanSettings>>('/renewals/settings', payload);
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

export function exportNonRenewalPlan(planId: string) {
  window.open(`/api/v1/renewals/plans/${planId}/non-renewal/export`, '_blank');
}

export async function listRenewalUnitPrices() {
  const { data } = await http.get<ApiResp<ListData<RenewalUnitPrice>>>('/renewals/unit-prices');
  return data;
}

export async function updateRenewalUnitPrices(prices: RenewalUnitPrice[]) {
  const { data } = await http.put<ApiResp<ListData<RenewalUnitPrice>>>('/renewals/unit-prices', { prices });
  return data;
}

export interface ContractPayload {
  contract_name: string;
  period_start: string;
  period_end: string;
  pre_tax_amount: number;
  supplier: string;
  business_contact: string;
  tech_contact: string;
}

export async function listContracts() {
  const { data } = await http.get<ApiResp<ListData<Contract>>>('/contracts');
  return data;
}

export async function createContract(payload: ContractPayload) {
  const { data } = await http.post<ApiResp<Contract>>('/contracts', payload);
  return data;
}

export async function updateContract(contractId: string, payload: ContractPayload) {
  const { data } = await http.put<ApiResp<Contract>>(`/contracts/${contractId}`, payload);
  return data;
}

export async function deleteContract(contractId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; contract_id: string }>>(`/contracts/${contractId}`);
  return data;
}

export async function uploadContractAttachment(contractId: string, file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<{ contract: Contract; attachment: ContractAttachment }>>(`/contracts/${contractId}/attachments`, form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export function downloadContractAttachment(contractId: string, attachmentId: string) {
  window.open(`/api/v1/contracts/${contractId}/attachments/${attachmentId}/download`, '_blank');
}

export async function deleteContractAttachment(contractId: string, attachmentId: string) {
  const { data } = await http.delete<ApiResp<Contract>>(`/contracts/${contractId}/attachments/${attachmentId}`);
  return data;
}


export async function listSuppliers(query?: string) {
  const params = query?.trim() ? { q: query.trim() } : {};
  const { data } = await http.get<ApiResp<ListData<Supplier>>>('/suppliers', { params });
  return data;
}

export async function createSupplier(payload: Omit<Supplier, 'supplier_id' | 'created_at' | 'updated_at'>) {
  const { data } = await http.post<ApiResp<Supplier>>('/suppliers', payload);
  return data;
}

export async function updateSupplier(supplierId: string, payload: Omit<Supplier, 'supplier_id' | 'created_at' | 'updated_at'>) {
  const { data } = await http.put<ApiResp<Supplier>>(`/suppliers/${supplierId}`, payload);
  return data;
}

export async function deleteSupplier(supplierId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; supplier_id: string }>>(`/suppliers/${supplierId}`);
  return data;
}


export interface SupplierImportFailure {
  row: number;
  reason: string;
}

export interface SupplierImportResult {
  created: number;
  updated: number;
  failed: number;
  failures?: SupplierImportFailure[];
}

export async function importSuppliers(file: File) {
  const form = new FormData();
  form.append('file', file);
  const { data } = await http.post<ApiResp<SupplierImportResult>>('/suppliers/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export function exportSuppliers() {
  window.open('/api/v1/suppliers/export', '_blank');
}

export function exportSupplierTemplate() {
  window.open('/api/v1/suppliers/template/export', '_blank');
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


export interface MetaModelPayload {
  model_code: string;
  model_name: string;
  description?: string;
}

export interface MetaModelUpdatePayload {
  model_name: string;
  description?: string;
}

export interface MetaFieldPayload {
  field_code: string;
  field_name: string;
  category?: string;
  value_type: string;
  required?: boolean;
  unique?: boolean;
  filterable?: boolean;
  sortable?: boolean;
  visible?: boolean;
  default_value?: string;
  validation_rule?: string;
  enum_options?: MetaEnumOption[];
}

export interface MetaFieldUpdatePayload extends Omit<MetaFieldPayload, 'field_code'> {}

export interface MetaReferencePayload {
  source_field_id: string;
  target_model_id: string;
  target_field_id: string;
  display_fields?: string[];
  on_delete_action?: string;
}

export async function listMetaModels(status?: string) {
  const { data } = await http.get<ApiResp<ListData<MetaModel>>>('/meta/models', { params: status ? { status } : undefined });
  return data;
}

export async function createMetaModel(payload: MetaModelPayload) {
  const { data } = await http.post<ApiResp<MetaModel>>('/meta/models', payload);
  return data;
}

export async function getMetaModel(modelId: string) {
  const { data } = await http.get<ApiResp<{ model: MetaModel; fields: MetaField[] }>>(`/meta/models/${modelId}`);
  return data;
}

export async function updateMetaModel(modelId: string, payload: MetaModelUpdatePayload) {
  const { data } = await http.put<ApiResp<MetaModel>>(`/meta/models/${modelId}`, payload);
  return data;
}

export async function archiveMetaModel(modelId: string) {
  const { data } = await http.post<ApiResp<MetaModel>>(`/meta/models/${modelId}/archive`, {});
  return data;
}

export async function deleteMetaModel(modelId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; model_id: string }>>(`/meta/models/${modelId}`);
  return data;
}

export async function cloneMetaModel(modelId: string, payload: { model_code: string; model_name: string; description?: string }) {
  const { data } = await http.post<ApiResp<MetaModel>>(`/meta/models/${modelId}/clone`, payload);
  return data;
}

export async function createMetaField(modelId: string, payload: MetaFieldPayload) {
  const { data } = await http.post<ApiResp<MetaField>>(`/meta/models/${modelId}/fields`, payload);
  return data;
}

export async function updateMetaField(modelId: string, fieldId: string, payload: MetaFieldUpdatePayload) {
  const { data } = await http.put<ApiResp<MetaField>>(`/meta/models/${modelId}/fields/${fieldId}`, payload);
  return data;
}

export async function deleteMetaField(modelId: string, fieldId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; field_id: string }>>(`/meta/models/${modelId}/fields/${fieldId}`);
  return data;
}

export async function reorderMetaFields(modelId: string, fieldIds: string[]) {
  const { data } = await http.put<ApiResp<MetaField[]>>(`/meta/models/${modelId}/fields/reorder`, { field_ids: fieldIds });
  return data;
}

export async function listMetaReferences(modelId: string) {
  const { data } = await http.get<ApiResp<ListData<MetaReference>>>(`/meta/models/${modelId}/references`);
  return data;
}

export async function createMetaReference(modelId: string, payload: MetaReferencePayload) {
  const { data } = await http.post<ApiResp<MetaReference>>(`/meta/models/${modelId}/references`, payload);
  return data;
}

export async function updateMetaReference(modelId: string, refId: string, payload: MetaReferencePayload) {
  const { data } = await http.put<ApiResp<MetaReference>>(`/meta/models/${modelId}/references/${refId}`, payload);
  return data;
}

export async function deleteMetaReference(modelId: string, refId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; ref_id: string }>>(`/meta/models/${modelId}/references/${refId}`);
  return data;
}


export async function publishMetaModel(modelId: string, payload?: { change_summary?: string; published_by?: string }) {
  const { data } = await http.post<ApiResp<MetaModelVersion>>(`/meta/models/${modelId}/publish`, payload || {});
  return data;
}

export async function listMetaModelVersions(modelId: string) {
  const { data } = await http.get<ApiResp<ListData<MetaModelVersion>>>(`/meta/models/${modelId}/versions`);
  return data;
}

export async function getMetaModelVersion(modelId: string, version: number) {
  const { data } = await http.get<ApiResp<{ version: MetaModelVersion; snapshot: MetaModelSnapshot }>>(`/meta/models/${modelId}/versions/${version}`);
  return data;
}

export async function rollbackMetaModel(modelId: string, version: number) {
  const { data } = await http.post<ApiResp<MetaModel>>(`/meta/models/${modelId}/rollback`, { version });
  return data;
}

export async function listMetaRecords(modelId: string) {
  const { data } = await http.get<ApiResp<{ model: MetaModel; fields: MetaField[]; list: MetaRecord[]; total: number }>>(`/meta/models/${modelId}/records`);
  return data;
}

export async function createMetaRecord(modelId: string, payload: { data: Record<string, any> }) {
  const { data } = await http.post<ApiResp<MetaRecord>>(`/meta/models/${modelId}/records`, payload);
  return data;
}

export async function updateMetaRecord(modelId: string, recordId: string, payload: { data: Record<string, any> }) {
  const { data } = await http.put<ApiResp<MetaRecord>>(`/meta/models/${modelId}/records/${recordId}`, payload);
  return data;
}

export async function deleteMetaRecord(modelId: string, recordId: string) {
  const { data } = await http.delete<ApiResp<{ deleted: boolean; record_id: string }>>(`/meta/models/${modelId}/records/${recordId}`);
  return data;
}

export function exportMetaRecordTemplate(modelId: string) {
  window.open(`/api/v1/meta/models/${modelId}/records/template/export`, '_blank');
}

export async function importMetaRecords(modelId: string, file: File, uniqueMode: 'strict' | 'off' = 'strict', importStrategy: 'append' | 'overwrite_all' = 'append') {
  const form = new FormData();
  form.append('file', file);
  form.append('unique_mode', uniqueMode);
  form.append('import_strategy', importStrategy);
  const { data } = await http.post<ApiResp<{ job_id: string; status: string; total: number }>>(`/meta/models/${modelId}/records/import`, form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
  return data;
}

export async function getMetaImportJob(jobId: string) {
  const { data } = await http.get<ApiResp<{ job_id: string; model_id: string; status: string; total: number; processed: number; success: number; failed: number; errors: any[]; message?: string }>>(`/meta/import-jobs/${jobId}`);
  return data;
}

export function exportMetaImportErrorsCSV(jobId: string) {
  window.open(`/api/v1/meta/import-jobs/${jobId}/errors/export`, '_blank');
}
