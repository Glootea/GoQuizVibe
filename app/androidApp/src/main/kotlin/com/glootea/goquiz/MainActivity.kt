package com.glootea.goquiz

import android.app.Application
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.glootea.goquiz.core.di.AppGraph
import com.glootea.goquiz.core.di.AppGraphFactory
import com.glootea.goquiz.feature.auth.data.KSafeFactory

class GoQuizApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        KSafeFactory.setAndroidContext(this)
    }
}

class MainActivity : ComponentActivity() {

    private val graph: AppGraph by lazy { AppGraphFactory.create() }

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)
        setContent {
            AppContent(graph)
        }
    }
}
