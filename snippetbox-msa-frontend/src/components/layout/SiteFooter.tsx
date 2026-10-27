export function SiteFooter() {
  return (
    <footer className="min-h-[60px] border-t border-line bg-surface-muted py-4 text-center text-ink-muted">
      Powered by <a href="https://echo.labstack.com/">Echo</a> and{' '}
      <a href="https://react.dev/">React</a> in {new Date().getFullYear()}
    </footer>
  )
}
