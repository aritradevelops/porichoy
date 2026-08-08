// Third-party brand marks for social login buttons. Kept as inline SVG (not an
// icon-font/CDN dependency) — official multi-color Google mark, monochrome Apple
// mark (matches Apple's own usage guidelines for light backgrounds).

export function GoogleIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" {...props}>
      <path
        fill="#4285F4"
        d="M23.52 12.27c0-.85-.08-1.67-.22-2.45H12v4.64h6.47a5.54 5.54 0 0 1-2.4 3.63v3h3.88c2.27-2.09 3.57-5.17 3.57-8.82Z"
      />
      <path
        fill="#34A853"
        d="M12 24c3.24 0 5.96-1.07 7.95-2.91l-3.88-3c-1.08.72-2.46 1.15-4.07 1.15-3.13 0-5.78-2.11-6.73-4.96H1.26v3.11A11.998 11.998 0 0 0 12 24Z"
      />
      <path
        fill="#FBBC05"
        d="M5.27 14.28A7.2 7.2 0 0 1 4.89 12c0-.79.14-1.56.38-2.28V6.61H1.26A11.998 11.998 0 0 0 0 12c0 1.94.46 3.77 1.26 5.39l4.01-3.11Z"
      />
      <path
        fill="#EA4335"
        d="M12 4.77c1.76 0 3.34.61 4.58 1.79l3.44-3.44C17.95 1.19 15.24 0 12 0 7.31 0 3.26 2.69 1.26 6.61l4.01 3.11C6.22 6.87 8.87 4.77 12 4.77Z"
      />
    </svg>
  )
}

export function AppleIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="16"
      height="16"
      aria-hidden="true"
      fill="currentColor"
      {...props}
    >
      <path d="M16.365 1.43c0 1.14-.468 2.096-1.14 2.844-.795.877-1.985 1.554-2.99 1.554-.09-.036-.16-.098-.16-.196a5.35 5.35 0 0 1 1.243-3.135C14.05.876 15.29.164 16.196.02c.11.196.169.395.169.41ZM20.63 17.44c-.585 1.35-.865 1.955-1.617 3.153-1.05 1.674-2.53 3.76-4.365 3.775-1.63.015-2.05-1.06-4.263-1.05-2.213.015-2.674 1.07-4.303 1.055-1.836-.015-3.235-1.9-4.286-3.573C-1.056 16.85-1.36 12.03.55 9.06c1.35-2.09 3.485-3.313 5.492-3.313 2.043 0 3.328 1.113 5.02 1.113 1.643 0 2.64-1.114 5.02-1.114 1.79 0 3.685.975 5.036 2.657-4.428 2.428-3.711 8.75.512 9.038Z" />
    </svg>
  )
}
