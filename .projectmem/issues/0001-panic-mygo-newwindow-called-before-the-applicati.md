# #0001 panic: mygo: NewWindow called before the application is ready; create windows in App.WhenReady

- 2026-11-03T11:44:14Z `issue`: panic: mygo: NewWindow called before the application is ready; create windows in App.WhenReady [main.go:36]
- 2026-11-03T11:44:32Z `attempt`: Move NewWindow call and event bindings inside mygo.App.WhenReady callback [main.go:34] (worked)
- 2026-11-03T11:44:57Z `fix`: Instantiate window inside mygo.App.WhenReady callback [main.go:37]
