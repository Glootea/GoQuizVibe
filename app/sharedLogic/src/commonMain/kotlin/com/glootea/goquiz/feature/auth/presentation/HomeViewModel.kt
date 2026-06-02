package com.glootea.goquiz.feature.auth.presentation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.glootea.goquiz.feature.auth.domain.model.User
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
class HomeViewModel(
    private val stateHolder: AuthStateHolder
) : ViewModel() {

    private val _state: MutableStateFlow<HomeState> = MutableStateFlow(HomeState())
    val state: StateFlow<HomeState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            stateHolder.state.collect { auth ->
                when (auth) {
                    is AuthState.Authenticated -> applyUser(auth.user)
                    else -> Unit
                }
            }
        }
    }

    private fun applyUser(user: User) {
        _state.update { it.copy(userName = user.name, userEmail = user.email) }
    }

    fun logout() {
        _state.update { it.copy(isLoggingOut = true) }
        stateHolder.performLogout()
    }
}
