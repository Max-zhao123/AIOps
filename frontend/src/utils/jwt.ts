export function parseJwt(token: string): { exp: number } | null {
  try {
    const base64 = token.split('.')[1]
    const payload = JSON.parse(atob(base64))
    return payload
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const payload = parseJwt(token)
  if (!payload || !payload.exp) return true
  // 提前 5 分钟视为过期
  return payload.exp * 1000 < Date.now() + 5 * 60 * 1000
}
