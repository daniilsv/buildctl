import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import BuildDetail from "./pages/BuildDetail";
import Dashboard from "./pages/Dashboard";
import Login from "./pages/Login";
import Callback from "./pages/Callback";
import ProjectDetail from "./pages/ProjectDetail";
import BranchDetail from "./pages/BranchDetail";
import Settings from "./pages/Settings";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/auth/login" element={<Login />} />
        <Route path="/auth/callback" element={<Callback />} />
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="projects/:id" element={<ProjectDetail />} />
          <Route path="projects/:id/branches/:branchName" element={<BranchDetail />} />
          <Route path="builds/:id" element={<BuildDetail />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
