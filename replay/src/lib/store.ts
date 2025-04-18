import { writable } from 'svelte/store'
import { getEvents } from './replay'

const events = getEvents()
export const currentIndex = writable(0)
export const currentEvent = writable(events[0])
export const allEvents = writable(events)

currentIndex.subscribe(i => currentEvent.set(events[i]))
