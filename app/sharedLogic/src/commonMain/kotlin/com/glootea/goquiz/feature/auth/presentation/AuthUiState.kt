package com.glootea.goquiz.feature.auth.presentation

data class LoginState(
    val email: String = "",
    val password: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
) {
    val isSubmitEnabled: Boolean
        get() = !isLoading && email.isNotBlank() && password.isNotEmpty()
}

data class RegisterState(
    val name: String = "",
    val email: String = "",
    val password: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
) {
    val isSubmitEnabled: Boolean
        get() = !isLoading && name.isNotBlank() && email.isNotBlank() && password.length >= 6
}

data class HomeState(
    val userName: String = "",
    val userEmail: String = "",
    val isLoggingOut: Boolean = false
)
