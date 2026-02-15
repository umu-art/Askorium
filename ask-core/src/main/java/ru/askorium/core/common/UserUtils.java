package ru.askorium.core.common;

import lombok.AccessLevel;
import lombok.NoArgsConstructor;
import org.springframework.web.context.request.RequestAttributes;
import org.springframework.web.context.request.RequestContextHolder;

import java.util.Objects;
import java.util.UUID;

@NoArgsConstructor(access = AccessLevel.PRIVATE)
public class UserUtils {

    public static final String USER_ID_KEY = "user_id_key";

    public static UUID getUserId() {
        return (UUID) Objects.requireNonNull(
                RequestContextHolder.currentRequestAttributes()
                        .getAttribute(USER_ID_KEY, RequestAttributes.SCOPE_REQUEST));
    }

}
