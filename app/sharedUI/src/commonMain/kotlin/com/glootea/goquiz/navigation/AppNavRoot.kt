package com.glootea.goquiz.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.remember
import androidx.navigation3.runtime.NavEntry
import androidx.navigation3.ui.NavDisplay
import com.glootea.goquiz.LocalAppGraph
import com.glootea.goquiz.feature.auth.presentation.AuthState
import com.glootea.goquiz.feature.auth.ui.home.HomeScreen
import com.glootea.goquiz.feature.auth.ui.landing.LandingScreen
import com.glootea.goquiz.feature.auth.ui.login.LoginScreen
import com.glootea.goquiz.feature.auth.ui.register.RegisterScreen

sealed class AppNavKey {
    data object Landing : AppNavKey()
    data object Login : AppNavKey()
    data object Register : AppNavKey()
    data object Home : AppNavKey()
}

@Composable
fun AppNavRoot() {
    val graph = LocalAppGraph.current
    val authState by graph.authStateHolder.state.collectAsState()

    val backStack = remember { mutableStateListOf<AppNavKey>(AppNavKey.Landing) }

    LaunchedEffect(authState) {
        when (authState) {
            is AuthState.Authenticated -> {
                backStack.clear()
                backStack.add(AppNavKey.Home)
            }
            AuthState.Unauthenticated -> {
                if (backStack.lastOrNull() == AppNavKey.Home) {
                    backStack.clear()
                    backStack.add(AppNavKey.Landing)
                }
            }
            AuthState.Unknown -> Unit
        }
    }

    NavDisplay(
        backStack = backStack,
        onBack = { backStack.removeLastOrNull() },
        entryProvider = { key ->
            when (key) {
                AppNavKey.Landing -> NavEntry(key) {
                    LandingScreen(
                        onLogin = { backStack.add(AppNavKey.Login) },
                        onRegister = { backStack.add(AppNavKey.Register) }
                    )
                }
                AppNavKey.Login -> NavEntry(key) {
                    LoginScreen(
                        onBack = { backStack.removeLastOrNull() },
                        onRegisterClick = { backStack.add(AppNavKey.Register) }
                    )
                }
                AppNavKey.Register -> NavEntry(key) {
                    RegisterScreen(
                        onBack = { backStack.removeLastOrNull() },
                        onLoginClick = { backStack.add(AppNavKey.Login) }
                    )
                }
                AppNavKey.Home -> NavEntry(key) {
                    HomeScreen()
                }
            }
        }
    )
}
