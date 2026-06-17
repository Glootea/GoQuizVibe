package com.glootea.goquiz.feature.auth.presentation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import com.glootea.goquiz.core.di.AppScope
import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@SingleIn(AppScope::class)
@Inject
class LoginViewModel(
    private val authRepository: AuthRepository,
    private val stateHolder: AuthStateHolder
) : ViewModel() {

    val state: StateFlow<LoginState>
        field: MutableStateFlow<LoginState> = MutableStateFlow(LoginState.Editing())

    fun onEmailChange(value: String) = state.update { current ->
        when (current) {
            is LoginState.Editing -> current.copy(email = value)
            is LoginState.Submitting -> current.copy(email = value)
            is LoginState.Error -> current.copy(email = value)
        }
    }

    fun onPasswordChange(value: String) = state.update { current ->
        when (current) {
            is LoginState.Editing -> current.copy(password = value)
            is LoginState.Submitting -> current.copy(password = value)
            is LoginState.Error -> current.copy(password = value)
        }
    }

    fun submit() {
        val current = state.value
        if (current !is LoginState.Editing || !current.isSubmitEnabled) return
        state.value = LoginState.Submitting(current.email, current.password)
        viewModelScope.launch {
            runCatching { authRepository.login(current.email.trim(), current.password) }
                .onSuccess { user ->
                    stateHolder.setAuthenticated(user)
                    state.value = LoginState.Editing()
                }
                .onFailure { e ->
                    state.value = LoginState.Error(
                        email = current.email,
                        password = current.password,
                        message = e.message ?: "Login failed"
                    )
                }
        }
    }
}
