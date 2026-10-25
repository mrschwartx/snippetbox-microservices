import { Link } from 'react-router-dom'
import { formatDateTime } from '@/lib/formatDate'
import type { Snippet } from '../types'

const CELL_CLASS = 'px-[1rem] py-2 text-left'
const MUTED_CELL_CLASS = `${CELL_CLASS} text-right text-ink-muted`

interface SnippetTableProps {
  snippets: Snippet[]
}

export function SnippetTable({ snippets }: SnippetTableProps) {
  return (
    <table className="w-full border-collapse border border-line bg-surface">
      <thead>
        <tr className="border-b border-line">
          <th className={CELL_CLASS}>Title</th>
          <th className={CELL_CLASS}>Created</th>
          <th className={MUTED_CELL_CLASS}>ID</th>
        </tr>
      </thead>
      <tbody>
        {snippets.map((snippet) => (
          <tr key={snippet.id} className="border-b border-line even:bg-surface-muted">
            <td className={CELL_CLASS}>
              <Link to={`/snippet/view/${snippet.id}`}>{snippet.title}</Link>
            </td>
            <td className={CELL_CLASS}>{formatDateTime(snippet.created)}</td>
            <td className={MUTED_CELL_CLASS}>#{snippet.id}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
