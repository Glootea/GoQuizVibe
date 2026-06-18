package com.glootea.goquiz.feature.auth.presentation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.core.di.AppScope
import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@SingleIn(AppScope::class)
@Inject
class HomeViewModel(
    private val stateHolder: AuthStateHolder
) : ViewModel() {

    val state: StateFlow<HomeState>
        field: MutableStateFlow<HomeState> = MutableStateFlow(HomeState())

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
        state.update { it.copy(userName = user.name, userEmail = user.email) }
    }

    fun logout() {
        state.update { it.copy(isLoggingOut = true) }
        stateHolder.performLogout()
    }
}
