import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './layout';

const ImportPage = lazy(() => import('./pages/ImportPage'));
const ContractPage = lazy(() => import('./pages/ContractPage'));
const PlanPage = lazy(() => import('./pages/PlanPage'));
const PlanDetailPage = lazy(() => import('./pages/PlanDetailPage'));
const FailureAnalysisPage = lazy(() => import('./pages/FailureAnalysisPage'));
const FailureDashboardPage = lazy(() => import('./pages/FailureDashboardPage'));

export default function App() {
  return (
    <Suspense fallback={null}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<Navigate to="/import" replace />} />
          <Route path="/import" element={<ImportPage />} />
          <Route path="/contracts" element={<ContractPage />} />
          <Route path="/plan" element={<PlanPage />} />
          <Route path="/plan/:planId" element={<PlanDetailPage />} />
          <Route path="/failure" element={<FailureAnalysisPage />} />
          <Route path="/failure/dashboard" element={<FailureDashboardPage />} />
          <Route path="/result" element={<Navigate to="/plan" replace />} />
          <Route path="/result/:planId" element={<PlanDetailPage />} />
        </Route>
      </Routes>
    </Suspense>
  );
}
