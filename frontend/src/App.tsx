import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './layout';
import ImportPage from './pages/ImportPage';
import PlanPage from './pages/PlanPage';
import ResultPage from './pages/ResultPage';

export default function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/import" replace />} />
        <Route path="/import" element={<ImportPage />} />
        <Route path="/plan" element={<PlanPage />} />
        <Route path="/result" element={<ResultPage />} />
        <Route path="/result/:planId" element={<ResultPage />} />
      </Route>
    </Routes>
  );
}
