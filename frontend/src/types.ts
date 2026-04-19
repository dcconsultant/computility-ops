export interface ApiResp<T> {
  code: number;
  message: string;
  data: T;
}

export interface ImportErrorItem {
  row: number;
  reason: string;
}

export interface ImportResult {
  total: number;
  success: number;
  failed: number;
  errors: ImportErrorItem[];
}

export interface ListData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface ServerItem {
  sn: string;
  manufacturer?: string;
  model?: string;
  psa: string;
  idc?: string;
  environment?: string;
  config_type: string;
  config_type_standardized?: string;
  package_standardized?: string;
  package_standardized_matched?: boolean;
  warranty_end_date?: string;
  launch_date?: string;
}

export interface HostPackageConfig {
  config_type: string;
  scene_category?: string;
  cpu_logical_cores: number;
  gpu_card_count?: number;
  data_disk_type?: string;
  data_disk_count?: number;
  storage_capacity_tb?: number;
  server_value_score?: number;
  arch_standardized_factor: number;
}

export interface SpecialRule {
  sn: string;
  manufacturer?: string;
  model?: string;
  psa?: string;
  idc?: string;
  package_type?: string;
  warranty_end_date?: string;
  launch_date?: string;
  policy: 'whitelist' | 'blacklist';
  reason?: string;
}

export interface ModelFailureRate {
  manufacturer: string;
  model: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface PackageFailureRate {
  period?: string;
  year?: number;
  config_type: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface PackageModelFailureRate {
  period?: string;
  year?: number;
  config_type: string;
  manufacturer: string;
  model: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface FailureRateSummary {
  period: 'history' | 'year' | string;
  year?: number;
  scope: 'all' | 'product' | 'devtest' | string;
  segment: 'storage' | 'non_storage' | string;
  full_cycle_failure_rate: number;
  over_warranty_failure_rate: number;
  fault_count: number;
  over_warranty_fault_count: number;
  server_years: number;
  over_warranty_years: number;
}

export interface FailureOverviewCard {
  segment: 'storage' | 'non_storage' | string;
  year: number;
  current_year_fault_rate: number;
  history_avg_fault_rate: number;
  current_year_fault_count: number;
  current_year_denominator: number;
  history_fault_count: number;
  history_denominator: number;
}

export interface FailureAgeTrendPoint {
  segment: 'storage' | 'non_storage' | string;
  age_bucket: number;
  numerator_fault_count: number;
  denominator_exposure: number;
  fault_rate: number;
}

export interface ImportErrorInsight {
  time: string;
  request_id: string;
  action: string;
  reason: string;
  hint: string;
}

export interface FailureFeatureFact {
  record_year_index: number;
  record_year_start: string;
  record_year_end: string;
  scope: string;
  scene_group: string;
  age_bucket: number;
  denominator_weighted: number;
  fault_count: number;
  fault_rate: number;
}

export interface StorageTopServerRate {
  sn: string;
  manufacturer?: string;
  model?: string;
  config_type?: string;
  environment?: string;
  idc?: string;
  warranty_end_date?: string;
  data_disk_count: number;
  single_disk_capacity_tb: number;
  total_capacity_tb: number;
  fault_count: number;
  denominator: number;
  fault_rate: number;
}

export type StorageBucket = 'warm_storage' | 'hot_storage';

export interface FaultYearAnalysisRow {
  row_no: number;
  sn?: string;
  created_at?: string;
  scope?: string;
  segment?: string;
  matched: boolean;
  remark?: string;
}

export interface FaultAnalysisResult {
  total_fault_rows: number;
  matched_fault_rows: number;
  generated_model_rates: number;
  generated_package_rates: number;
  generated_package_model_rates: number;
  overall_rates?: FailureRateSummary[];
  failure_feature_facts?: FailureFeatureFact[];
  storage_top_server_rates?: StorageTopServerRate[];
  year_fault_analysis_rows?: FaultYearAnalysisRow[];
}

export interface PlanItem {
  rank: number;
  bucket?: string;
  sn: string;
  manufacturer?: string;
  model?: string;
  environment?: string;
  config_type: string;
  scene_category?: string;
  cpu_logical_cores: number;
  gpu_card_count?: number;
  storage_capacity_tb?: number;
  psa: string;
  arch_standardized_factor: number;
  base_score?: number;
  afr_old?: number;
  afr_avg?: number;
  failure_adjust_factor?: number;
  final_score: number;
  special_policy?: string;
}

export interface RenewalPlanSection {
  bucket: 'compute' | 'warm_storage' | 'hot_storage' | 'gpu' | string;
  target_cores?: number;
  target_storage_tb?: number;
  covered_cores?: number;
  covered_storage_tb?: number;
  required_cores?: number;
  required_storage_tb?: number;
  covered_count?: number;
  selected_cores?: number;
  selected_storage_tb?: number;
  selected_count: number;
  items: PlanItem[];
}

export interface NonRenewalItem {
  sn: string;
  bucket?: string;
  manufacturer?: string;
  model?: string;
  environment?: string;
  config_type?: string;
  psa?: string;
  final_score?: number;
  reason_code: string;
  reason: string;
  reason_detail?: string;
  rank_in_bucket?: number;
}

export interface RenewalPlan {
  plan_id: string;
  target_date?: string;
  excluded_environments?: string[];
  excluded_psas?: string[];
  target_cores: number;
  warm_target_storage_tb?: number;
  hot_target_storage_tb?: number;
  covered_compute_cores?: number;
  covered_warm_storage_tb?: number;
  covered_hot_storage_tb?: number;
  required_compute_cores?: number;
  required_warm_storage_tb?: number;
  required_hot_storage_tb?: number;
  unmatched_config_count?: number;
  unmatched_config_types?: string[];
  gpu_current_cards?: number;
  gpu_current_servers?: number;
  gpu_covered_cards?: number;
  gpu_covered_servers?: number;
  gpu_renewal_cards?: number;
  gpu_renewal_servers?: number;
  selected_cores: number;
  selected_storage_tb?: number;
  selected_count: number;
  items: PlanItem[];
  non_renewal_items?: NonRenewalItem[];
  sections?: RenewalPlanSection[];
}
