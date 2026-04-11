import { FileText } from 'lucide-react'

export default function ReportList() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Reports</h1>

      <div className="bg-white rounded-lg shadow">
        <div className="text-center py-12 text-gray-500">
          <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No reports yet</p>
          <p className="text-sm mt-2">Reports will appear here after case execution</p>
        </div>
      </div>
    </div>
  )
}
