import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './layout';

const ImportPage = lazy(() => import('./pages/ImportPage'));
const ContractPage = lazy(() => import('./pages/ContractPage'));
const SupplierPage = lazy(() => import('./pages/SupplierPage'));
const DeliveryTrackingPage = lazy(() => import('./pages/DeliveryTrackingPage'));
const PlanPage = lazy(() => import('./pages/PlanPage'));
const PlanDetailPage = lazy(() => import('./pages/PlanDetailPage'));
const FailureAnalysisPage = lazy(() => import('./pages/FailureAnalysisPage'));
const FailureDashboardPage = lazy(() => import('./pages/FailureDashboardPage'));
const MetaModelPage = lazy(() => import('./pages/MetaModelPage'));
const ResourcePlanningPage = lazy(() => import('./pages/ResourcePlanningPage'));
const ReconfigManagementPage = lazy(() => import('./pages/ReconfigManagementPage'));

export default function App() {
  return (
    <Suspense fallback={null}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<Navigate to="/import" replace />} />
          <Route path="/import" element={<ImportPage />} />
          <Route path="/resource-planning" element={<ResourcePlanningPage />} />
          <Route path="/contracts" element={<ContractPage />} />
          <Route path="/suppliers" element={<SupplierPage />} />
          <Route path="/delivery-tracking" element={<DeliveryTrackingPage />} />
          <Route path="/plan" element={<PlanPage />} />
          <Route path="/plan/:planId" element={<PlanDetailPage />} />
          <Route path="/failure" element={<FailureAnalysisPage />} />
          <Route path="/failure/dashboard" element={<FailureDashboardPage />} />
          <Route path="/reconfig" element={<ReconfigManagementPage />} />
          <Route path="/meta-models" element={<MetaModelPage />} />
          <Route path="/result" element={<Navigate to="/plan" replace />} />
          <Route path="/result/:planId" element={<PlanDetailPage />} />
        </Route>
      </Routes>
    </Suspense>
  );
}
