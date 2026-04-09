import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './layout';
import ImportPage from './pages/ImportPage';
import PlanPage from './pages/PlanPage';
export default function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/import" replace />} />
        <Route path="/import" element={<ImportPage />} />
        <Route path="/plan" element={<PlanPage />} />
        <Route path="/result" element={<Navigate to="/plan" replace />} />
        <Route path="/result/:planId" element={<Navigate to="/plan" replace />} />
      </Route>
    </Routes>
  );
}
