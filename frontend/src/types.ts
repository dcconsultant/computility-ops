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
  detailed_config?: string;
  psa: string;
  idc?: string;
  environment?: string;
  config_type: string;
  config_type_standardized?: string;
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
  power_watts?: number;
  release_year?: number;
  memory_capacity_gb?: number;
  arch_standardized_factor: number;
}

export interface CabinetUtilizationSetting {
  utilization: number;
}

export interface CabinetConfig {
  id: number;
  idc: string;
  rated_power_kw: number;
  monthly_rent: number;
}

export interface ValueScoreCabinetBaseline {
  status: string;
  idc: string;
  cabinet_utilization: number;
  min_rated_power_kw: number;
  monthly_rent_cny: number;
  formula: string;
  source_count: number;
  note?: string;
}

export interface ValueScoreCostParams {
  depreciation_months: number;
  network_device_share_cny: number;
  server_renewal_fee_cny: number;
  cabinet_utilization: number;
  rated_power_kw: number;
  monthly_rent_cny: number;
}

export interface ValueScoreOriginalValue {
  config_type: string;
  server_original_cny: number;
}

export interface ValueScorePerformanceParam {
  config_type: string;
  unavailable_cores: number;
  unavailable_memory_gb: number;
  performance_score: number;
}

export interface ValueScorePerformanceAlert {
  config_type: string;
  error_code: string;
  field: string;
  current_value: string;
  suggestion: string;
}

export interface ValueScorePerformanceCalcItem {
  config_type: string;
  cpu_logical_cores: number;
  memory_capacity_gb: number;
  unavailable_cores: number;
  unavailable_memory_gb: number;
  performance_score: number;
  available_cores: number;
  available_memory_gb: number;
  standard_score: number;
  cpu_performance_factor: number;
  memory_ratio: number;
  memory_ratio_factor: number;
  overall_performance_ratio: number;
  alerts?: ValueScorePerformanceAlert[];
}

export interface ValueScorePerformanceCalcResult {
  items: ValueScorePerformanceCalcItem[];
  alert_count: number;
  note?: string;
}

export interface ValueScoreTCOItem {
  config_type: string;
  power_watts: number;
  power_kw: number;
  cabinet_cost_monthly: number;
  server_original_cny: number;
  depreciation_monthly: number;
  network_device_monthly: number;
  network_cabinet_monthly: number;
  server_renewal_monthly: number;
  other_fixed_cost_monthly: number;
  total_tco_monthly: number;
}

export interface ValueScoreTCOResult {
  status: string;
  idc: string;
  cabinet_utilization: number;
  min_rated_power_kw: number;
  monthly_rent_cny: number;
  depreciation_months: number;
  formula: string;
  items: ValueScoreTCOItem[];
  note?: string;
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
  recent_1y_failure_rate?: number;
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
  detailed_config?: string;
  environment?: string;
  idc?: string;
  config_type: string;
  scene_category?: string;
  cpu_logical_cores: number;
  gpu_card_count?: number;
  storage_capacity_tb?: number;
  recent_1y_failure_rate?: number;
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
  idc?: string;
  config_type?: string;
  psa?: string;
  final_score?: number;
  reason_code: string;
  reason: string;
  reason_detail?: string;
  rank_in_bucket?: number;
}

export type RenewalTargetMode = 'manual' | 'maximize';

export interface RenewalSceneTarget {
  mode: RenewalTargetMode;
  target: number;
  min_performance_score?: number;
  min_single_disk_capacity_tb?: number;
}

export interface RenewalRegionTargets {
  compute: RenewalSceneTarget;
  warm_storage: RenewalSceneTarget;
  hot_storage: RenewalSceneTarget;
  gpu: RenewalSceneTarget;
}

export interface RenewalRequirements {
  domestic: RenewalRegionTargets;
  india: RenewalRegionTargets;
}

export interface RenewalPlanSettings {
  target_date: string;
  excluded_environments: string[];
  excluded_psas: string[];
  idle_stopped_psas?: string[];
  requirements: RenewalRequirements;
  domestic_budget: number;
  india_budget: number;
}

export interface RenewalUnitPrice {
  country: string;
  scene_category: 'compute' | 'warm_storage' | 'hot_storage' | 'gpu' | string;
  unit_price: number;
}

export interface RenewalPlan {
  plan_id: string;
  target_date?: string;
  excluded_environments?: string[];
  excluded_psas?: string[];
  idle_stopped_psas?: string[];
  target_cores: number;
  warm_target_storage_tb?: number;
  hot_target_storage_tb?: number;
  domestic_budget?: number;
  india_budget?: number;
  total_servers_no_psa?: number;
  domestic_servers_no_psa?: number;
  india_servers_no_psa?: number;
  requirements?: RenewalRequirements;
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
  full_renewal?: RenewalPlanVariant;
  minimal_renewal?: RenewalPlanVariant;
  minimal_renewal_error?: string;
  comparison?: RenewalComparison;
}

export interface MinimalComputeStats {
  failure_rate_year: number;
  failure_rate: number;
  domestic_current_cores: number;
  domestic_idle_stopped_cores: number;
  domestic_reserve_cores: number;
  domestic_guarantee_cores: number;
  domestic_minimal_renew_cores: number;
  india_current_cores: number;
  india_idle_stopped_cores: number;
  india_reserve_cores: number;
  india_guarantee_cores: number;
  india_minimal_renew_cores: number;
  total_current_cores: number;
  total_idle_stopped_cores: number;
  total_reserve_cores: number;
  total_guarantee_cores: number;
  total_minimal_renew_cores: number;
}

export interface RenewalPlanVariant {
  name: string;
  target_cores: number;
  required_compute_cores?: number;
  covered_compute_cores?: number;
  selected_cores: number;
  selected_storage_tb?: number;
  selected_count: number;
  gpu_renewal_cards?: number;
  gpu_renewal_servers?: number;
  items: PlanItem[];
  non_renewal_items?: NonRenewalItem[];
  sections?: RenewalPlanSection[];
  minimal_compute_metrics?: MinimalComputeStats;
}

export interface ReducedRenewalItem {
  sn: string;
  idc?: string;
  psa?: string;
  config_type?: string;
  cpu_logical_cores: number;
  full_rank?: number;
  reason: string;
  saved_amount: number;
}

export interface RenewalComparison {
  full_renewal_count: number;
  minimal_renewal_count: number;
  reduced_count: number;
  full_amount: number;
  minimal_amount: number;
  saved_amount: number;
  saved_ratio: number;
  full_renewal_cores: number;
  minimal_renewal_cores: number;
  reduced_renewal_cores: number;
  reduced_renewal_items?: ReducedRenewalItem[];
}

export interface ContractAttachment {
  attachment_id: string;
  file_name: string;
  file_size: number;
  mime_type?: string;
  uploaded_at: string;
}

export interface Contract {
  contract_id: string;
  contract_name: string;
  period_start: string;
  period_end: string;
  pre_tax_amount: number;
  supplier: string;
  business_contact: string;
  tech_contact: string;
  attachments?: ContractAttachment[];
  created_at?: string;
  updated_at?: string;
}

export interface ArrivalPlan {
  plan_id: string;
  category: '服务器' | '网络设备' | '耗材及配件' | string;
  material_code: string;
  material_name: string;
  quantity: number;
  receiving_address: string;
  supplier: string;
  order_no: string;
  asset_code_range: string;
  estimated_arrival_time: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
}

export interface DeviceArrivalRecord {
  record_id: string;
  category: '服务器' | '网络设备' | string;
  package_code: string;
  package_type: string;
  material_service_code: string;
  material_service_description: string;
  rack_units: number;
  manufacturer: string;
  quantity: number;
  receiving_location: string;
  purchase_request_no: string;
  srm_requirement_submitted_at: string;
  po_no: string;
  actual_arrival_time: string;
  created_at?: string;
  updated_at?: string;
}

export interface AccessoryArrivalRecord {
  record_id: string;
  purchase_request_no: string;
  material_service_code: string;
  material_service_description: string;
  quantity: number;
  supplier: string;
  idc_room: string;
  purchase_background: string;
  srm_requirement_submitted_at: string;
  po_no: string;
  arrival_time: string;
  created_at?: string;
  updated_at?: string;
}


export type MetaModelStatus = 'draft' | 'published' | 'archived';

export interface MetaModel {
  id: string;
  model_code: string;
  model_name: string;
  description?: string;
  status: MetaModelStatus;
  current_version: number;
  created_at: string;
  updated_at: string;
}

export interface MetaEnumOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface MetaField {
  id: string;
  model_id: string;
  field_code: string;
  field_name: string;
  category?: string;
  value_type: string;
  required: boolean;
  unique: boolean;
  filterable: boolean;
  sortable: boolean;
  visible: boolean;
  default_value?: string;
  validation_rule?: string;
  enum_options?: MetaEnumOption[];
  sort_no: number;
  created_at: string;
  updated_at: string;
}

export interface MetaReference {
  id: string;
  model_id: string;
  source_field_id: string;
  target_model_id: string;
  target_field_id: string;
  display_fields: string[];
  on_delete_action: string;
  created_at: string;
  updated_at: string;
}


export interface MetaModelVersion {
  id: string;
  model_id: string;
  version_no: number;
  snapshot_json: string;
  published_at: string;
  published_by?: string;
  change_summary?: string;
}

export interface MetaModelSnapshot {
  model: MetaModel;
  fields: MetaField[];
  references: MetaReference[];
}

export interface MetaRecord {
  id: string;
  model_id: string;
  data: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface ResourcePlanningRequest {
  self_build_value_score_compute?: number;
  self_build_value_score_warm_storage?: number;
  self_build_value_score_hot_storage?: number;
  self_build_value_score_gpu?: number;
  public_cloud_value_score_compute?: number;
  public_cloud_value_score_warm_storage?: number;
  public_cloud_value_score_hot_storage?: number;
  public_cloud_value_score_gpu?: number;
  compute_demand_cores: number;
  warm_storage_demand_tb: number;
  hot_storage_demand_tb: number;
  gpu_demand_cards: number;
  cabinet_and_other_cost_cny: number;
  annual_depreciation_cny: number;
  disposal_psas: string;
  non_business_psas: string;
  reconfig_compute_server_count?: number;
  reconfig_compute_capacity?: number;
  reconfig_compute_cost?: number;
  reconfig_warm_server_count?: number;
  reconfig_warm_capacity?: number;
  reconfig_warm_cost?: number;
  reconfig_hot_server_count?: number;
  reconfig_hot_capacity?: number;
  reconfig_hot_cost?: number;
  reconfig_gpu_server_count?: number;
  reconfig_gpu_capacity?: number;
  reconfig_gpu_cost?: number;
  reconfig_done_server_count?: number;
  reconfig_done_logical_cores?: number;
  reconfig_done_warm_storage_tb?: number;
  reconfig_done_hot_storage_tb?: number;
  reconfig_done_gpu_cards?: number;
  reconfig_done_cost_cny?: number;
  quasi_compute_server_count?: number;
  quasi_compute_capacity?: number;
  quasi_compute_cost?: number;
  quasi_warm_server_count?: number;
  quasi_warm_capacity?: number;
  quasi_warm_cost?: number;
  quasi_hot_server_count?: number;
  quasi_hot_capacity?: number;
  quasi_hot_cost?: number;
  quasi_gpu_server_count?: number;
  quasi_gpu_capacity?: number;
  quasi_gpu_cost?: number;
  quasi_purchase_server_count: number;
  quasi_purchase_logical_cores: number;
  quasi_purchase_warm_storage_tb?: number;
  quasi_purchase_hot_storage_tb?: number;
  quasi_purchase_gpu_cards?: number;
  quasi_purchase_cost_cny: number;
  executed_new_compute_server_count?: number;
  executed_new_compute_capacity?: number;
  executed_new_compute_cost?: number;
  executed_new_warm_server_count?: number;
  executed_new_warm_capacity?: number;
  executed_new_warm_cost?: number;
  executed_new_hot_server_count?: number;
  executed_new_hot_capacity?: number;
  executed_new_hot_cost?: number;
  executed_new_gpu_server_count?: number;
  executed_new_gpu_capacity?: number;
  executed_new_gpu_cost?: number;
  executed_new_purchase_server_count?: number;
  executed_new_purchase_logical_cores?: number;
  executed_new_purchase_warm_storage_tb?: number;
  executed_new_purchase_hot_storage_tb?: number;
  executed_new_purchase_gpu_cards?: number;
  executed_new_purchase_cost_cny?: number;
}

export interface ResourcePlanningConfigState {
  saved_at: string;
  config: ResourcePlanningRequest;
}

export interface ResourcePlanningScenePurchasePlan {
  scene_category: string;
  package_config_type: string;
  package_release_year: number;
  server_count: number;
  covered_logical_cores?: number;
  covered_storage_tb?: number;
  covered_gpu_cards?: number;
  purchase_amount_cny: number;
  annual_cost_cny: number;
  annual_budget_cny: number;
  value_score?: number;
}

export interface ResourcePlanningResponse {
  generated_at: string;
  config: ResourcePlanningRequest;
  reconfig_plan: {
    source_plan_id: string;
    server_count: number;
    logical_cores: number;
    covered_warm_storage_tb: number;
    covered_hot_storage_tb: number;
    covered_gpu_cards: number;
    cost_cny: number;
  };
  quasi_purchase_plan: {
    server_count: number;
    logical_cores: number;
    covered_warm_storage_tb: number;
    covered_hot_storage_tb: number;
    covered_gpu_cards: number;
    cost_cny: number;
  };
  new_purchase_plan: {
    package_config_type: string;
    package_release_year: number;
    server_count: number;
    covered_logical_cores: number;
    covered_warm_storage_tb: number;
    covered_hot_storage_tb: number;
    covered_gpu_cards: number;
    base_demand_cores: number;
    routine_replacement_cores: number;
    extra_replacement_cores: number;
    total_replacement_cores: number;
    purchase_amount_cny: number;
    annual_cost_cny: number;
    annual_budget_cny: number;
    value_score: number;
    scene_plans?: ResourcePlanningScenePurchasePlan[];
  };
  renewal_plan: {
    source_plan_id: string;
    device_count: number;
    covered_compute_cores: number;
    covered_warm_storage_tb: number;
    covered_hot_storage_tb: number;
    covered_gpu_cards: number;
    budget_cny: number;
  };
  self_repair_plan: {
    device_count: number;
    covered_cores: number;
  };
  disposal_plan: {
    device_count: number;
    covered_compute_cores: number;
    covered_warm_storage_tb: number;
    covered_hot_storage_tb: number;
    covered_gpu_cards: number;
    unmatched_package_count: number;
    matched_psa_server_count: number;
    normalized_psas: string[];
  };
  result_analysis: {
    amount: {
      reconfig_cost_cny: number;
      quasi_purchase_cost_cny: number;
      new_purchase_cost_cny: number;
      renewal_cost_cny: number;
      cabinet_other_cost_cny: number;
      total_cost_cny: number;
    };
    cost: {
      reconfig_cost_cny: number;
      quasi_purchase_cost_cny: number;
      new_purchase_cost_cny: number;
      renewal_cost_cny: number;
      depreciation_cost_cny: number;
      cabinet_other_cost_cny: number;
      total_cost_cny: number;
    };
    compute_capacity: {
      reconfig_cores: number;
      quasi_purchase_cores: number;
      new_purchase_cores: number;
      stock_continue_cores: number;
      total_cores: number;
    };
    warm_storage_capacity: {
      reconfig_tb: number;
      quasi_purchase_tb: number;
      new_purchase_tb: number;
      stock_continue_tb: number;
      total_tb: number;
    };
    hot_storage_capacity: {
      reconfig_tb: number;
      quasi_purchase_tb: number;
      new_purchase_tb: number;
      stock_continue_tb: number;
      total_tb: number;
    };
    gpu_capacity: {
      reconfig_cards: number;
      quasi_purchase_cards: number;
      new_purchase_cards: number;
      stock_continue_cards: number;
      total_cards: number;
    };
    non_business_psas?: string[];
    available_compute_cores?: number;
    available_warm_storage_tb?: number;
    available_hot_storage_tb?: number;
    available_gpu_cards?: number;
  };
}

export interface DeliveryDecisionInput {
  hw_total: number;
  hw_cores: number;
  hw_tax_rate: number;
  idc_rent_monthly: number;
  idc_rack_kw: number;
  idc_fill_rate: number;
  idc_server_power_w: number;
  idc_network_depreciation: number;
  cloud_memory_ratio: number;
  cloud_disk_ratio: number;
  cloud_cpu_daily_price: number;
  cloud_memory_daily_price: number;
  cloud_disk_daily_price: number;
  cloud_tax_rate: number;
  depreciation_years: number;
  wacc_rate: number;
  residual_rate: number;
  country: string;
  currency: string;
  cloud_current_discount: number;
}

export interface DeliveryDecisionDefaults {
  country: string;
  currency: string;
  input: DeliveryDecisionInput;
}

export interface DeliveryDecisionFormulaTrace {
  cloud_gross: number;
  cloud_daily_net: number;
  hardware_net_per_core: number;
  server_kw: number;
  physical_monthly_net: number;
  physical_daily_net: number;
  daily_depreciation: number;
  daily_wacc: number;
  daily_depreciation_3y: number;
  self_daily_tco: number;
  self_daily_tco_3y: number;
  premium_ratio_r: number;
  self_weight: number;
  cloud_weight: number;
  formula_self_share: number;
  formula_cloud_share: number;
  final_self_share: number;
  final_cloud_share: number;
  daily_margin: number;
  break_even_years?: number | null;
  cloud_hedge_lost: boolean;
  equality_tolerance: number;
}

export interface DeliveryDecisionSensitivityPoint {
  curve: 'hardware_price' | 'cloud_discount';
  label: string;
  x_value: number;
  cloud_daily_net: number;
  self_daily_tco: number;
  self_daily_tco_3y: number;
  formula_self_share: number;
  final_self_share: number;
  cloud_hedge_lost: boolean;
}

export interface DeliveryDecisionSnapshot {
  operator: string;
  formula_version: string;
  calculated_at: string;
}

export interface DeliveryDecisionResult {
  input: DeliveryDecisionInput;
  formula: DeliveryDecisionFormulaTrace;
  sensitivity_points: DeliveryDecisionSensitivityPoint[];
  snapshot: DeliveryDecisionSnapshot;
}

export interface Supplier {
  supplier_id: string;
  company_full_name: string;
  tax_number: string;
  project_owner: string;
  project_owner_phone: string;
  tech_contact: string;
  tech_contact_phone: string;
  business_scope: string;
  created_at?: string;
  updated_at?: string;
}
