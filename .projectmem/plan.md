# godraw — plan

> Editable **intent** file: ideas + plans — what we *mean to do*.
> This is NOT the event log. `events.jsonl` -> `summary.md` records what
> *happened*; this file records what we *intend*. The AI reads it at
> session start and edits it directly (like `PROJECT_MAP.md`): add ideas
> and plans, check items off, move done work down to Shipped. Plans are
> never logged as events.

## Ideas
_Loose thoughts, not yet committed to._

## Active plans

## Next
_Queued, but not started._

## Someday / maybe

## Shipped
_Move completed plans here so the top stays about the future._
- [x] Implement central library storage manager in `~/Documents/godraw/`
- [x] Create Home Page Gallery view in `internal/app/` (grid of cards, thumbnail preview, "+ New Drawing" button)
- [x] Add "+ New Drawing" modal prompt for file name before canvas creation
- [x] Add card actions for Rename and Delete (with confirmation modal)
- [x] Wire single-window view router in `App` (`ViewHome` ⇄ `ViewCanvas`) with "← Home" button
- [x] Auto-generate/cache thumbnails on drawing save
- [x] Handle CLI arguments by importing external `.tldr` files into `~/Documents/godraw/`
