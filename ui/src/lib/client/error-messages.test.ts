import { describe, expect, it } from 'vitest'
import { resolveErrorMessage } from './error-messages'

describe('resolveErrorMessage', () => {
  it('maps a known key to its local message', () => {
    expect(resolveErrorMessage('identity.invalid_credentials', 'backend text')).toBe(
      'Incorrect email or password.',
    )
  })

  it('falls back to the backend-provided message for an unknown key', () => {
    expect(resolveErrorMessage('some.new.key', 'backend text')).toBe('backend text')
  })
})
