import { describe, expect, it } from 'vitest'
import { hasPermission } from './permissions'

describe('hasPermission', () => {
  it('matches an exact module:action@scope entry', () => {
    expect(hasPermission(['tenants:list@tenant'], 'tenants', 'list')).toBe(true)
  })

  it('matches a wildcard action', () => {
    expect(hasPermission(['tenants:*@root'], 'tenants', 'list')).toBe(true)
  })

  it('does not match a different module', () => {
    expect(hasPermission(['domains:*@root'], 'tenants', 'list')).toBe(false)
  })

  it('does not match a different action', () => {
    expect(hasPermission(['tenants:create@tenant'], 'tenants', 'list')).toBe(false)
  })

  it('ignores a malformed entry with no scope', () => {
    expect(hasPermission(['tenants:list'], 'tenants', 'list')).toBe(false)
  })

  it('returns false for an empty permission list', () => {
    expect(hasPermission([], 'tenants', 'list')).toBe(false)
  })
})
