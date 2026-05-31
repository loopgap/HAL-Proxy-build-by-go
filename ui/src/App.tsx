import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import { ErrorBoundary } from './components/ui/ErrorBoundary'
import { Breadcrumbs, useBreadcrumbs } from './components/ui/Breadcrumbs'
import { AuthGuard } from './guards/AuthGuard'
import NotFoundPage from './pages/NotFound'
import Login from './pages/Login'
import LoadingSpinner from './components/ui/LoadingSpinner'

const Dashboard = lazy(() => import('./pages/Dashboard'))
const CaseList = lazy(() => import('./pages/CaseList'))
const CaseCreate = lazy(() => import('./pages/CaseCreate'))
const CaseDetail = lazy(() => import('./pages/CaseDetail'))
const ApprovalList = lazy(() => import('./pages/ApprovalList'))
const ReportList = lazy(() => import('./pages/ReportList'))

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-64">
      <LoadingSpinner size="lg" label="Loading page..." />
    </div>
  )
}

function PageWrapper({ children }: { children: React.ReactNode }) {
  const breadcrumbs = useBreadcrumbs()
  return (
    <div className="space-y-4">
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
          <Route path="/login" element={<Login />} />
          <Route
            path="/"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <Dashboard />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path="/cases"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <CaseList />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path="/cases/new"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <CaseCreate />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path="/cases/:id"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <CaseDetail />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path="/approvals"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <ApprovalList />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route
            path="/reports"
            element={
              <AuthGuard>
                <ErrorBoundary>
                  <Suspense fallback={<PageLoader />}>
                    <ReportList />
                  </Suspense>
                </ErrorBoundary>
              </AuthGuard>
            }
          />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </PageWrapper>
    </Layout>
  )
}

export default App
