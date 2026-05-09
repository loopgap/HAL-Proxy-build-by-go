import { lazy } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import { RouteErrorBoundary } from './components/ui/RouteErrorBoundary'
import { Breadcrumbs, useBreadcrumbs } from './components/ui/Breadcrumbs'
import { AuthGuard } from './guards/AuthGuard'
import NotFoundPage from './pages/NotFound'
import Login from './pages/Login'

// Lazy load all page components for code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'))
const CaseList = lazy(() => import('./pages/CaseList'))
const CaseDetail = lazy(() => import('./pages/CaseDetail'))
const ApprovalList = lazy(() => import('./pages/ApprovalList'))
const ReportList = lazy(() => import('./pages/ReportList'))

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
        <Routes>
          <Route path='/login' element={<Login />} />
          <Route
            path='/'
            element={
              <AuthGuard>
                <RouteErrorBoundary>
                  <Dashboard />
                </RouteErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path='/cases'
            element={
              <AuthGuard>
                <RouteErrorBoundary>
                  <CaseList />
                </RouteErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path='/cases/:id'
            element={
              <AuthGuard>
                <RouteErrorBoundary>
                  <CaseDetail />
                </RouteErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path='/approvals'
            element={
              <AuthGuard>
                <RouteErrorBoundary>
                  <ApprovalList />
                </RouteErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path='/reports'
            element={
              <AuthGuard>
                <RouteErrorBoundary>
                  <ReportList />
                </RouteErrorBoundary>
              </AuthGuard>
            }
          />
          <Route path='*' element={<NotFoundPage />} />
        </Routes>
      </PageWrapper>
    </Layout>
  )
}

export default App