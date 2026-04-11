import { CheckCircle, AlertTriangle } from 'lucide-react'

export default function ApprovalList() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Approvals</h1>

      <div className="bg-white rounded-lg shadow">
        <div className="text-center py-12 text-gray-500">
          <CheckCircle className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No pending approvals</p>
          <p className="text-sm mt-2">Approvals will appear here when case execution requires authorization</p>
        </div>
      </div>
    </div>
  )
}
