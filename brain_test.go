package joe

// TODO: test shutdown event context is not already canceled → Brain test
// TODO: test NewBrain uses in memory brain by default
// TODO: test RegisterHandler
//       → simple
//       → simple with error
//       → simple with context
//       → simple with context and error
//       → error cases
// TODO: test Brain.Emit is asynchronous
// TODO: test HandleEvents
//       → InitEvent
//       → multiple handlers can match
//       → no handlers can match (e.g. wrong EventType)
//       → callbacks
//       → timeouts
//       → context done and shutdown event
// TODO: BrainMemoryEvents
