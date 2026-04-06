const statusConfig = {
  operational: { label: 'Operational', bg: 'bg-emerald-50', text: 'text-emerald-700', dot: 'bg-emerald-500' },
  degraded: { label: 'Degraded', bg: 'bg-yellow-50', text: 'text-yellow-700', dot: 'bg-yellow-500' },
  major_outage: { label: 'Major Outage', bg: 'bg-red-50', text: 'text-red-700', dot: 'bg-red-500' },
  maintenance: { label: 'Maintenance', bg: 'bg-blue-50', text: 'text-blue-700', dot: 'bg-blue-500' },
  // incident statuses
  investigating: { label: 'Investigating', bg: 'bg-orange-50', text: 'text-orange-700', dot: 'bg-orange-500' },
  identified: { label: 'Identified', bg: 'bg-yellow-50', text: 'text-yellow-700', dot: 'bg-yellow-500' },
  monitoring: { label: 'Monitoring', bg: 'bg-blue-50', text: 'text-blue-700', dot: 'bg-blue-500' },
  resolved: { label: 'Resolved', bg: 'bg-emerald-50', text: 'text-emerald-700', dot: 'bg-emerald-500' },
}

const impactConfig = {
  none: { label: 'None', bg: 'bg-gray-50', text: 'text-gray-600' },
  minor: { label: 'Minor', bg: 'bg-yellow-50', text: 'text-yellow-700' },
  major: { label: 'Major', bg: 'bg-orange-50', text: 'text-orange-700' },
  critical: { label: 'Critical', bg: 'bg-red-50', text: 'text-red-700' },
}

export function StatusBadge({ status }) {
  const config = statusConfig[status] || { label: status, bg: 'bg-gray-50', text: 'text-gray-600', dot: 'bg-gray-400' }
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${config.dot}`} />
      {config.label}
    </span>
  )
}

export function ImpactBadge({ impact }) {
  const config = impactConfig[impact] || { label: impact, bg: 'bg-gray-50', text: 'text-gray-600' }
  return (
    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      {config.label}
    </span>
  )
}
