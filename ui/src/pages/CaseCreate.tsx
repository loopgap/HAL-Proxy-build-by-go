import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { useCreateCase } from '@/hooks/useApi'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import type { RiskClass } from '@/types'

export default function CaseCreate() {
  const navigate = useNavigate()
  const createCaseMutation = useCreateCase()
  const [title, setTitle] = useState('')
  const [command, setCommand] = useState('')
  const [riskClass, setRiskClass] = useState<RiskClass>('observe')
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!title.trim()) {
      setError('Title is required')
      return
    }

    try {
      await createCaseMutation.mutateAsync({
        title: title.trim(),
        commands: [
          {
            name: 'Run Command',
            action: command.trim() || 'echo "hello"',
            risk_class: riskClass,
          },
        ],
      })
      navigate('/cases')
    } catch (err: any) {
      setError(err?.message || 'Failed to create case')
    }
  }

  return (
    <div className="max-w-xl mx-auto">
      <div className="flex items-center gap-4 mb-6">
        <Link to="/cases" className="text-gray-500 hover:text-gray-700">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-bold">New Case</h1>
      </div>

      <Card>
        <CardHeader
          title="Create New Case"
          subtitle="Define a new test execution unit with high-risk control parameters."
        />
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm p-3 rounded-md">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <label
                htmlFor="title"
                className="text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Title
              </label>
              <input
                id="title"
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white"
                placeholder="Enter case title"
              />
            </div>

            <div className="space-y-2">
              <label
                htmlFor="command"
                className="text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Command
              </label>
              <textarea
                id="command"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                className="w-full min-h-24 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white"
                placeholder="Enter command to run (e.g. ls -la)"
              />
            </div>

            <div className="space-y-2">
              <label
                htmlFor="riskClass"
                className="text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Risk Class
              </label>
              <select
                id="riskClass"
                value={riskClass}
                onChange={(e) => setRiskClass(e.target.value as RiskClass)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white font-sans"
              >
                <option value="observe">Observe (Low Risk - Read Only)</option>
                <option value="mutate">Mutate (Medium Risk - State Change)</option>
                <option value="destructive">Destructive (High Risk - Requires Approval)</option>
                <option value="exclusive">Exclusive (Critical Risk - Exclusive Lock)</option>
              </select>
            </div>

            <div className="flex gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <Button type="submit" className="flex-1" disabled={createCaseMutation.isPending}>
                {createCaseMutation.isPending ? 'Creating...' : 'Create Case'}
              </Button>
              <Link to="/cases" className="flex-1">
                <Button type="button" variant="outline" className="w-full">
                  Cancel
                </Button>
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
