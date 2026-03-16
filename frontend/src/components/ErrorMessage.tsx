interface Props {
  message: string
}

export default function ErrorMessage({ message }: Props) {
  return (
    <div className="rounded-md bg-red-950 border border-red-800 px-4 py-3 text-sm text-red-300">
      {message}
    </div>
  )
}
