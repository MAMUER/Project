import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { store } from './store/store';

const App: React.FC = () => {
    return (
        <Provider store={store}>
            <BrowserRouter>
                <Routes>
                    <Route path="/" element={<Navigate to="/dashboard" />} />
                    <Route path="/dashboard" element={<div>Dashboard Page</div>} />
                    <Route path="/login" element={<div>Login Page</div>} />
                    <Route path="/profile" element={<div>Profile Page</div>} />
                    <Route path="/training" element={<div>Training Page</div>} />
                </Routes>
            </BrowserRouter>
        </Provider>
    );
};

export default App;