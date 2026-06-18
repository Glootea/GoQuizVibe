import SwiftUI
import SharedLogic

@main
struct iOSApp: App {
    @StateObject private var coordinator = AppCoordinator()

    var body: some Scene {
        WindowGroup {
            RootView(coordinator: coordinator)
        }
    }
}

final class AppCoordinator: ObservableObject {
    let graph: AppGraph = AppGraphFactory().create()
    @Published var state: AuthState = .Unknown()
    private var watchTask: Task<Void, Never>?

    init() {
        watchTask = Task { [weak self] in
            guard let stream = self?.graph.authStateHolder.state else { return }
            for await s in stream {
                await MainActor.run { self?.state = s }
            }
        }
    }

    deinit {
        watchTask?.cancel()
    }
}

struct RootView: View {
    @ObservedObject var coordinator: AppCoordinator

    var body: some View {
        switch coordinator.state {
        case .Unknown:
            VStack { ProgressView() }
        case .Unauthenticated:
            LandingView(
                onLogin: {},
                onRegister: {}
            )
        case .Authenticated:
            HomeView(coordinator: coordinator)
        }
    }
}

struct LandingView: View {
    let onLogin: () -> Void
    let onRegister: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Spacer()
            Text("GoQuizVibe")
                .font(.largeTitle)
                .foregroundColor(.indigo)
                .bold()
            Text("Образовательная платформа с тестами и учебными материалами")
                .multilineTextAlignment(.center)
                .foregroundColor(.secondary)
                .padding(.horizontal, 24)
            Spacer()
            Button(action: onLogin) {
                Text("Войти")
                    .frame(maxWidth: .infinity, minHeight: 44)
            }
            .buttonStyle(.borderedProminent)
            Button(action: onRegister) {
                Text("Регистрация")
                    .frame(maxWidth: .infinity, minHeight: 44)
            }
            .buttonStyle(.bordered)
            Spacer()
        }
        .padding(.horizontal, 24)
    }
}

struct HomeView: View {
    @ObservedObject var coordinator: AppCoordinator

    var body: some View {
        VStack(spacing: 16) {
            Spacer()
            Text("Добро пожаловать")
                .font(.title2)
            if case .authenticated(let user) = coordinator.state {
                Text(user.name)
                    .font(.largeTitle)
                    .bold()
                Text(user.email)
                    .font(.callout)
                    .foregroundColor(.secondary)
            }
            Spacer()
            Button("Выйти") {
                coordinator.graph.homeViewModel.logout()
            }
            .buttonStyle(.bordered)
        }
    }
}
