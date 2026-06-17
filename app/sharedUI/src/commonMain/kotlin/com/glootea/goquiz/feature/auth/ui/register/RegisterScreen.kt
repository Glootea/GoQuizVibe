package com.glootea.goquiz.feature.auth.ui.register

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.glootea.goquiz.LocalAppGraph
import com.glootea.goquiz.feature.auth.presentation.RegisterState
import com.glootea.goquiz.feature.auth.presentation.errorMessage
import com.glootea.goquiz.feature.auth.presentation.isSubmitEnabled
import com.glootea.goquiz.feature.auth.ui.components.AppTextField
import com.glootea.goquiz.feature.auth.ui.components.AuthFormScaffold
import com.glootea.goquiz.feature.auth.ui.components.PrimaryButton
import com.glootea.goquiz.theme.GoQuizTheme

@Composable
fun RegisterScreen(
    onBack: () -> Unit,
    onLoginClick: () -> Unit
) {
    val graph = LocalAppGraph.current
    val viewModel = remember(graph) { graph.registerViewModel }
    val state by viewModel.state.collectAsStateWithLifecycle()

    GoQuizTheme {
        AuthFormScaffold(
            title = "Регистрация",
            errorMessage = state.errorMessage
        ) {
            AppTextField(
                value = state.name,
                onValueChange = viewModel::onNameChange,
                label = "Имя"
            )
            Spacer(Modifier.height(12.dp))
            AppTextField(
                value = state.email,
                onValueChange = viewModel::onEmailChange,
                label = "Email",
                keyboardType = KeyboardType.Email
            )
            Spacer(Modifier.height(12.dp))
            AppTextField(
                value = state.password,
                onValueChange = viewModel::onPasswordChange,
                label = "Пароль (минимум 6)",
                isPassword = true
            )
            Spacer(Modifier.height(20.dp))
            PrimaryButton(
                text = "Создать аккаунт",
                onClick = viewModel::submit,
                enabled = state.isSubmitEnabled,
                isLoading = state is RegisterState.Submitting
            )
            Spacer(Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.Center,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Уже есть аккаунт?",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onBackground
                )
                TextButton(onClick = onLoginClick) {
                    Text("Войти")
                }
            }
        }
    }
}
