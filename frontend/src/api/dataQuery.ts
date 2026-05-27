import client from './client'
import type { DataQueryRequest, DataQueryResponse } from '@/types'

export async function queryData(data: DataQueryRequest): Promise<DataQueryResponse> {
  const res = await client.post('/data/query', data)
  return res.data?.data ?? res.data
}
