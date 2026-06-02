package com.glootea.goquiz.feature.auth.presentation

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.feature.auth.domain.usecase.LogoutUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.ObserveMeUseCase
import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

@SingleIn(AppScope::class)
@Inject
class AuthStateHolder(
    private val observeMe: ObserveMeUseCase,
    private val logout: LogoutUseCase,
    private val scope: CoroutineScope
) {
    private val _state: MutableStateFlow<AuthState> = MutableStateFlow(AuthState.Unknown)
    val state: StateFlow<AuthState> = _state.asStateFlow()

    init {
        scope.launch { refreshInternal() }
    }

    fun setAuthenticated(user: com.glootea.goquiz.feature.auth.domain.model.User) {
        _state.value = AuthState.Authenticated(user)
    }

    fun setUnauthenticated() {
        _state.value = AuthState.Unauthenticated
    }

    fun refresh() {
        scope.launch { refreshInternal() }
    }

    fun performLogout() {
        scope.launch {
            runCatching { logout() }
            _state.value = AuthState.Unauthenticated
        }
    }

    private suspend fun refreshInternal() {
        val user = runCatching { observeMe() }.getOrNull()
        _state.value = if (user != null) AuthState.Authenticated(user) else AuthState.Unauthenticated
    }
}
