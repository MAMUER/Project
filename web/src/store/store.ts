import { configureStore } from '@reduxjs/toolkit';
import { useDispatch, useSelector } from 'react-redux';

interface UserState {
    user: null | {
        id: string;
        email: string;
        role: string;
    };
    token: string | null;
    loading: boolean;
    error: string | null;
}

const initialState: UserState = {
    user: null,
    token: localStorage.getItem('token'),
    loading: false,
    error: null,
};

const userReducer = (state = initialState, action: any): UserState => {
    switch (action.type) {
        case 'LOGIN_SUCCESS':
            return {
                ...state,
                user: action.payload.user,
                token: action.payload.token,
                loading: false,
                error: null,
            };
        case 'LOGIN_FAILURE':
            return {
                ...state,
                loading: false,
                error: action.payload.error,
            };
        case 'LOGOUT':
            return {
                ...initialState,
                token: null,
            };
        default:
            return state;
    }
};

export const store = configureStore({
    reducer: {
        user: userReducer,
    },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();