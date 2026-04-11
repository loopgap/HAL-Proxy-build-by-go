import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import CaseList from './pages/CaseList'
import CaseDetail from './pages/CaseDetail'
import ApprovalList from './pages/ApprovalList'
import ReportList from './pages/ReportList'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/cases" element={<CaseList />} />
        <Route path="/cases/:id" element={<CaseDetail />} />
        <Route path="/approvals" element={<ApprovalList />} />
        <Route path="/reports" element={<ReportList />} />
      </Routes>
    </Layout>
  )
}

export default App
