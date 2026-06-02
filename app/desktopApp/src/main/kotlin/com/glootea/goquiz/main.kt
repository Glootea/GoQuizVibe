package com.glootea.goquiz

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application

fun main() = application {
    Window(
        onCloseRequest = ::exitApplication,
        title = "GoQuiz",
    ) {
        App()
    }
}