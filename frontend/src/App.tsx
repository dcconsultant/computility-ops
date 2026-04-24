import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './layout';
import ImportPage from './pages/ImportPage';
import PlanPage from './pages/PlanPage';
import PlanDetailPage from './pages/PlanDetailPage';
import FailureAnalysisPage from './pages/FailureAnalysisPage';
import FailureDashboardPage from './pages/FailureDashboardPage';
import ContractPage from './pages/ContractPage';

export default function App() {
  return (
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
  );
}
