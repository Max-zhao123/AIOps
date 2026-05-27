import client from './client'
import type { Inspection, InspectionReport } from '@/types'

export async function listInspections(): Promise<Inspection[]> {
  const res = await client.get('/inspections')
  return res.data?.data ?? res.data ?? []
}

export async function createInspection(data: Partial<Inspection>): Promise<Inspection> {
  const res = await client.post('/inspections', data)
  return res.data?.data ?? res.data
}

export async function runInspection(id: number): Promise<void> {
  await client.post(`/inspections/${id}/run`)
}

export async function listReports(inspectionId: number): Promise<InspectionReport[]> {
  const res = await client.get(`/inspections/${inspectionId}/reports`)
  return res.data?.data ?? res.data ?? []
}
