package com.glootea.goquiz.feature.auth.presentation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.glootea.goquiz.feature.auth.domain.usecase.LoginUseCase
import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import com.glootea.goquiz.core.di.AppScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@SingleIn(AppScope::class)
@Inject
class LoginViewModel(
    private val login: LoginUseCase,
    private val stateHolder: AuthStateHolder
) : ViewModel() {

    private val _state: MutableStateFlow<LoginState> = MutableStateFlow(LoginState())
    val state: StateFlow<LoginState> = _state.asStateFlow()

    fun onEmailChange(value: String) = _state.update { it.copy(email = value, error = null) }
    fun onPasswordChange(value: String) = _state.update { it.copy(password = value, error = null) }

    fun submit() {
        val current = _state.value
        if (!current.isSubmitEnabled) return
        _state.update { it.copy(isLoading = true, error = null) }
        viewModelScope.launch {
            runCatching { login(current.email, current.password) }
                .onSuccess { user ->
                    stateHolder.setAuthenticated(user)
                    _state.update { it.copy(isLoading = false, error = null) }
                }
                .onFailure { e ->
                    _state.update { it.copy(isLoading = false, error = e.message ?: "Login failed") }
                }
        }
    }
}
