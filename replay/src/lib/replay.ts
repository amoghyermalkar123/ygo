import type { Event } from './types'
import events from '../../events.json'

export function getEvents(): Event[] {
  return events
}
