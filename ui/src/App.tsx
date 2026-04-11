import { Suspense, lazy } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'

// Lazy load all page components for code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'))
const CaseList = lazy(() => import('./pages/CaseList'))
const CaseDetail = lazy(() => import('./pages/CaseDetail'))
const ApprovalList = lazy(() => import('./pages/ApprovalList'))
const ReportList = lazy(() => import('./pages/ReportList'))

// Loading fallback component
function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center h-64">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
    </div>
  )
}

function App() {
  return (
    <Layout>
      <Suspense fallback={<LoadingSpinner />}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/cases" element={<CaseList />} />
          <Route path="/cases/:id" element={<CaseDetail />} />
          <Route path="/approvals" element={<ApprovalList />} />
          <Route path="/reports" element={<ReportList />} />
        </Routes>
      </Suspense>
    </Layout>
  )
}

export default App
