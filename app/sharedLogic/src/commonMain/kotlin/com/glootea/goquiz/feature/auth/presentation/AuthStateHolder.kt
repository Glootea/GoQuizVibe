package com.glootea.goquiz.feature.auth.presentation

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

@SingleIn(AppScope::class)
@Inject
class AuthStateHolder(
    private val authRepository: AuthRepository,
    private val scope: CoroutineScope
) {


    val state: StateFlow<AuthState>
        field = MutableStateFlow<AuthState>(AuthState.Unknown)

    init {
        scope.launch { refreshInternal() }
    }

    fun setAuthenticated(user: com.glootea.goquiz.feature.auth.domain.model.User) {
        state.value = AuthState.Authenticated(user)
    }

    fun setUnauthenticated() {
        state.value = AuthState.Unauthenticated
    }

    fun refresh() {
        scope.launch { refreshInternal() }
    }

    fun performLogout() {
        scope.launch {
            runCatching { authRepository.logout() }
            state.value = AuthState.Unauthenticated
        }
    }

    private suspend fun refreshInternal() {
        val user = runCatching { authRepository.me() }.getOrNull()
        state.value = if (user != null) AuthState.Authenticated(user) else AuthState.Unauthenticated
    }
}
