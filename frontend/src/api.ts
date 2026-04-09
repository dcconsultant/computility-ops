import axios from 'axios';
import type {
  ApiResp,
  HostPackageConfig,
  ImportResult,
  ListData,
  ModelFailureRate,
  PackageFailureRate,
  PackageModelFailureRate,
  RenewalPlan,
  ServerItem,
  SpecialRule
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

export interface CreatePlanPayload {
  target_date: string;
  excluded_environments: string[];
  target_cores: number;
  warm_target_storage_tb: number;
  hot_target_storage_tb: number;
}

export async function createPlan(payload: CreatePlanPayload) {
  const { data } = await http.post<ApiResp<RenewalPlan>>('/renewals/plan', payload);
  return data;
}

export async function getPlan(planId: string) {
  const { data } = await http.get<ApiResp<RenewalPlan>>(`/renewals/plans/${planId}`);
  return data;
}

export function exportPlan(planId: string, format: 'xlsx' | 'csv') {
  window.open(`/api/v1/renewals/plans/${planId}/export?format=${format}`, '_blank');
}
