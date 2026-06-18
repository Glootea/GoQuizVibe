package com.glootea.goquiz

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import com.glootea.goquiz.core.di.AppGraphFactory

fun main() = application {
    val graph = AppGraphFactory.create()
    Window(
        onCloseRequest = ::exitApplication,
        title = "GoQuiz",
    ) {
        AppContent(graph)
    }
}
