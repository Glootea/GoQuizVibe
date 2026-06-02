package com.glootea.goquiz.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.remember
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

    val backstack = remember { mutableStateListOf<AppNavKey>(AppNavKey.Landing) }

    LaunchedEffect(authState) {
        when (authState) {
            is AuthState.Authenticated -> {
                backstack.clear()
                backstack.add(AppNavKey.Home)
            }
            AuthState.Unauthenticated -> {
                if (backstack.lastOrNull() == AppNavKey.Home) {
                    backstack.clear()
                    backstack.add(AppNavKey.Landing)
                }
            }
            AuthState.Unknown -> Unit
        }
    }

    when (val current = backstack.lastOrNull() ?: AppNavKey.Landing) {
        AppNavKey.Landing -> LandingScreen(
            onLogin = { backstack.add(AppNavKey.Login) },
            onRegister = { backstack.add(AppNavKey.Register) }
        )
        AppNavKey.Login -> LoginScreen(
            onBack = { backstack.removeLastOrNull() },
            onRegisterClick = { backstack.add(AppNavKey.Register) }
        )
        AppNavKey.Register -> RegisterScreen(
            onBack = { backstack.removeLastOrNull() },
            onLoginClick = { backstack.add(AppNavKey.Login) }
        )
        AppNavKey.Home -> HomeScreen()
    }
}
