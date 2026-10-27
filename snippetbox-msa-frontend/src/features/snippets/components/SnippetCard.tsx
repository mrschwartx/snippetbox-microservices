import { formatDateTime } from '@/lib/formatDate'
import type { Snippet } from '../types'

const META_CLASS =
  'flex flex-wrap justify-between gap-[1rem] bg-surface-muted px-[1rem] py-[0.75em] text-ink-muted'

interface SnippetCardProps {
  snippet: Snippet
}

export function SnippetCard({ snippet }: SnippetCardProps) {
  return (
    <article className="rounded-[3px] border border-line bg-surface">
      <div className={META_CLASS}>
        <strong className="text-ink">{snippet.title}</strong>
        <span>#{snippet.id}</span>
      </div>
      <pre className="overflow-x-auto border-y border-line p-[1rem]">
        <code className="whitespace-pre">{snippet.content}</code>
      </pre>
      <div className={META_CLASS}>
        <time dateTime={snippet.created}>Created: {formatDateTime(snippet.created)}</time>
        <time dateTime={snippet.expires}>Expires: {formatDateTime(snippet.expires)}</time>
      </div>
    </article>
  )
}
