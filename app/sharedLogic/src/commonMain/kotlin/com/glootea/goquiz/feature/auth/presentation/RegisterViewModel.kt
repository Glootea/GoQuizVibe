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
class RegisterViewModel(
    private val authRepository: AuthRepository,
    private val stateHolder: AuthStateHolder
) : ViewModel() {

    val state: StateFlow<RegisterState>
        field: MutableStateFlow<RegisterState> = MutableStateFlow(RegisterState.Editing())

    fun onNameChange(value: String) = state.update { current ->
        when (current) {
            is RegisterState.Editing -> current.copy(name = value)
            is RegisterState.Submitting -> current.copy(name = value)
            is RegisterState.Error -> current.copy(name = value)
        }
    }

    fun onEmailChange(value: String) = state.update { current ->
        when (current) {
            is RegisterState.Editing -> current.copy(email = value)
            is RegisterState.Submitting -> current.copy(email = value)
            is RegisterState.Error -> current.copy(email = value)
        }
    }

    fun onPasswordChange(value: String) = state.update { current ->
        when (current) {
            is RegisterState.Editing -> current.copy(password = value)
            is RegisterState.Submitting -> current.copy(password = value)
            is RegisterState.Error -> current.copy(password = value)
        }
    }

    fun submit() {
        val current = state.value
        if (current !is RegisterState.Editing || !current.isSubmitEnabled) return
        state.value = RegisterState.Submitting(current.name, current.email, current.password)
        viewModelScope.launch {
            runCatching {
                authRepository.register(current.name.trim(), current.email.trim(), current.password)
            }
                .onSuccess { user ->
                    stateHolder.setAuthenticated(user)
                    state.value = RegisterState.Editing()
                }
                .onFailure { e ->
                    state.value = RegisterState.Error(
                        name = current.name,
                        email = current.email,
                        password = current.password,
                        message = e.message ?: "Registration failed"
                    )
                }
        }
    }
}
