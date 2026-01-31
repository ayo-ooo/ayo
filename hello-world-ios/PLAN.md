# Hello World iOS App Plan

## Requirements Recap
- SwiftUI-based iOS application that displays a friendly greeting.
- Project must live entirely inside `./hello-world-ios`.
- Source must be idiomatic Swift + SwiftUI.
- App must build, test, and run in the iOS Simulator **via command line only**.
- Provide clear commands/scripts so others can reproduce build, test, and simulator run steps.

## Work Breakdown
1. **Project Skeleton**
   - Create a minimal Xcode project (`HelloWorldIOS.xcodeproj`) with an app target and a unit test target.
   - Configure iOS deployment target (e.g., iOS 17.0) and bundle identifier (e.g., `com.example.helloworldios`).
   - Ensure the scheme is shared so `xcodebuild` can reference it.

2. **SwiftUI Implementation**
   - Implement `HelloWorldIOSApp.swift` as the app entry point using `@main` and `WindowGroup`.
   - Implement `ContentView.swift` that renders a greeting using idiomatic SwiftUI, factoring out any view model/state so it can be unit tested cleanly.
   - Provide any supporting resources (e.g., Assets catalog) needed for a clean build.

3. **Unit Tests**
   - Build a lightweight view model or helper that exposes the greeting string.
   - Write XCTest cases under `HelloWorldIOSTests` that validate the greeting logic to ensure `xcodebuild test` has meaningful coverage.

4. **Command-Line Tooling**
   - Document exact `xcodebuild` commands for building and testing against the iOS Simulator (e.g., iPhone 15 / iOS 17.5).
   - Add a small automation wrapper (Makefile or shell script) to simplify `build`, `test`, and `sim` workflows from the CLI.
   - For “simulate,” compile for the simulator, install the `.app` onto a booted simulator device via `xcrun simctl`, and launch it, ensuring commands are idempotent.

5. **Verification & Documentation**
   - Run the documented commands to prove: build succeeds, tests pass, simulator launch works.
   - Capture concise instructions in `README.md` (within `hello-world-ios`) covering prerequisites and the CLI flow.

## Acceptance Criteria
- `xcodebuild -scheme HelloWorldIOS -destination 'platform=iOS Simulator,name=iPhone 15' build` succeeds.
- `xcodebuild ... test` succeeds with passing tests.
- A provided script/Make target boots (or reuses) a simulator, installs the freshly built app, and launches it without manual Xcode interaction.
- Documentation explains how to replicate the above.
