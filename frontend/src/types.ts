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
  warranty_end_date?: string;
  launch_date?: string;
}

export interface HostPackageConfig {
  config_type: string;
  scene_category?: string;
  cpu_logical_cores: number;
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
}

export interface ModelFailureRate {
  manufacturer: string;
  model: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface PackageFailureRate {
  config_type: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface PackageModelFailureRate {
  config_type: string;
  manufacturer: string;
  model: string;
  failure_rate: number;
  over_warranty_failure_rate?: number;
}

export interface FaultAnalysisResult {
  total_fault_rows: number;
  matched_fault_rows: number;
  generated_model_rates: number;
  generated_package_rates: number;
  generated_package_model_rates: number;
}

export interface PlanItem {
  rank: number;
  bucket?: string;
  sn: string;
  manufacturer?: string;
  model?: string;
  environment?: string;
  config_type: string;
  cpu_logical_cores: number;
  storage_capacity_tb?: number;
  psa: number;
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
  selected_cores?: number;
  selected_storage_tb?: number;
  selected_count: number;
  items: PlanItem[];
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
  selected_cores: number;
  selected_storage_tb?: number;
  selected_count: number;
  items: PlanItem[];
  sections?: RenewalPlanSection[];
}
