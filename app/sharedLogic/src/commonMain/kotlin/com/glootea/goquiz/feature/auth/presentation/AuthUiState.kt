package com.glootea.goquiz.feature.auth.presentation

sealed interface LoginState {
    val email: String
    val password: String

    data class Editing(
        override val email: String = "",
        override val password: String = ""
    ) : LoginState

    data class Submitting(
        override val email: String,
        override val password: String
    ) : LoginState

    data class Error(
        override val email: String,
        override val password: String,
        val message: String
    ) : LoginState
}

val LoginState.isSubmitEnabled: Boolean
    get() = this is LoginState.Editing &&
        email.isNotBlank() &&
        password.isNotEmpty()

val LoginState.errorMessage: String?
    get() = (this as? LoginState.Error)?.message

sealed interface RegisterState {
    val name: String
    val email: String
    val password: String

    data class Editing(
        override val name: String = "",
        override val email: String = "",
        override val password: String = ""
    ) : RegisterState

    data class Submitting(
        override val name: String,
        override val email: String,
        override val password: String
    ) : RegisterState

    data class Error(
        override val name: String,
        override val email: String,
        override val password: String,
        val message: String
    ) : RegisterState
}

val RegisterState.isSubmitEnabled: Boolean
    get() = this is RegisterState.Editing &&
        name.isNotBlank() &&
        email.isNotBlank() &&
        password.length >= MIN_PASSWORD_LENGTH

val RegisterState.errorMessage: String?
    get() = (this as? RegisterState.Error)?.message

const val MIN_PASSWORD_LENGTH: Int = 6

data class HomeState(
    val userName: String = "",
    val userEmail: String = "",
    val isLoggingOut: Boolean = false
)
