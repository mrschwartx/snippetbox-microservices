import { Link } from 'react-router-dom'

export function SiteHeader() {
  return (
    <header className="brand-bar border-b border-line pt-[33px] pb-[27px] text-center">
      <h1>
        <Link to="/" className="logo-link">
          SnippetBox
        </Link>
      </h1>
    </header>
  )
}
