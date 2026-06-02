package com.glootea.goquiz.feature.auth.presentation

import com.glootea.goquiz.feature.auth.domain.model.User

sealed class AuthState {
    data object Unknown : AuthState()
    data object Unauthenticated : AuthState()
    data class Authenticated(val user: User) : AuthState()
}
