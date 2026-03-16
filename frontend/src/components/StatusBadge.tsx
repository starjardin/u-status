interface Props {
  status: 'up' | 'down' | 'pending'
  size?: 'sm' | 'md'
}

const colors = {
  up: 'bg-green-500',
  down: 'bg-red-500',
  pending: 'bg-yellow-500',
}

const labels = {
  up: 'UP',
  down: 'DOWN',
  pending: 'PENDING',
}

export default function StatusBadge({ status, size = 'md' }: Props) {
  const dotSize = size === 'sm' ? 'w-2 h-2' : 'w-2.5 h-2.5'
  const textSize = size === 'sm' ? 'text-xs' : 'text-sm'

  return (
    <span className={`inline-flex items-center gap-1.5 font-medium ${textSize}`}>
      <span
        className={`${dotSize} rounded-full ${colors[status]} ${status === 'down' ? 'animate-pulse' : ''}`}
      />
      <span
        className={
          status === 'up'
            ? 'text-green-400'
            : status === 'down'
              ? 'text-red-400'
              : 'text-yellow-400'
        }
      >
        {labels[status]}
      </span>
    </span>
  )
}
