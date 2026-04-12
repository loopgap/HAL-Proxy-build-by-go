import { Suspense, lazy } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import { ErrorBoundary } from './components/ui/ErrorBoundary'
import { LoadingSpinner } from './components/ui/LoadingSpinner'
import { Breadcrumbs, useBreadcrumbs } from './components/ui/Breadcrumbs'
import NotFoundPage from './pages/NotFound'
import Login from './pages/Login'

// Lazy load all page components for code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'))
const CaseList = lazy(() => import('./pages/CaseList'))
const CaseDetail = lazy(() => import('./pages/CaseDetail'))
const ApprovalList = lazy(() => import('./pages/ApprovalList'))
const ReportList = lazy(() => import('./pages/ReportList'))

function PageLoader() {
  return (
    <div className='flex items-center justify-center h-64'>
      <LoadingSpinner size='lg' label='Loading page...' />
    </div>
  )
}

function PageWrapper({ children }: { children: React.ReactNode }) {
  const breadcrumbs = useBreadcrumbs()
  return (
    <div className='space-y-4'>
      {breadcrumbs.length > 0 && <Breadcrumbs items={breadcrumbs} />}
      {children}
    </div>
  )
}

function App() {
  return (
    <Layout>
      <PageWrapper>
        <ErrorBoundary>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path='/login' element={<Login />} />
              <Route path='/' element={<Dashboard />} />
              <Route path='/cases' element={<CaseList />} />
              <Route path='/cases/:id' element={<CaseDetail />} />
              <Route path='/approvals' element={<ApprovalList />} />
              <Route path='/reports' element={<ReportList />} />
              <Route path='*' element={<NotFoundPage />} />
            </Routes>
          </Suspense>
        </ErrorBoundary>
      </PageWrapper>
    </Layout>
  )
}

export default App