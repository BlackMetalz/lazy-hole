# Day 11: Feb 13, 2026

# Undo action
Goal: for easier life, support only 1 last action!

So little different to init

```go
// This need to init, because it will create new make, without init it will be nil --> panic
var effectTracker = NewEffectTracker()

// UndoStack
// No need constructor because inside it has only sync.Mutex + slice []UndoAction
// both 2 zero value can be use
// Mutex default unlocked
// Slice default nil
// So we don't need constructor, got it?
var undoStack = &UndoStack{}
```

Ok, why pointer? because method receiver is `(s *UndoStack)` - pointer receiver!. If `undoStack` is not pointer we will call `undoStack.Push()` will modify the fucking copy, not modify the original, got it?

And about `&UndoStack{}` is really confused me. I know it is related to the way how we define and called pointer. But not really complete understand like constructor example above. Holy fucking shit.

Alright, that is the fucking hardest part. Oh no, another shit going downn. Took me a while to figure out why Gemini generate 2 func `Peek` and `Pop`, normally when i write someshit related to CLI, not really need to have 2 func. I asked again and get explain: It is related to `tview async` not related to logic.

So `Peek()` is used to get last action, but not remove it from stack. `Pop()` is used to get last action and remove it from stack. Why need both, because `Peek()` get info but keep stack (no idea user yes/no) while `Pop()` remove it from stack (user confirm yes). That is why we need 2 separated func!

Summary 3 methods: `Push, Pop, Peek`
- `Push()`: User add effect success -> add action into stack
- `Peek()`: When user hit `u`, Look at `UndoAction` struct but not delete, for display pop up, "Undo this"?
- `Pop()`: User confirm yes -> remove last action from stack --> return it --> call removeSingleEffect() with it.