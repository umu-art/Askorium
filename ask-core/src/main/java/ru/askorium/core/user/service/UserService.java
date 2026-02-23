package ru.askorium.core.user.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.user.domain.UserEntity;
import ru.askorium.core.user.jpa.UserJpa;

import java.time.OffsetDateTime;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class UserService {

    private final UserJpa userJpa;

    @Transactional(transactionManager = "userTransactionManager")
    public UserResolution resolveOrCreate(UUID cookieUid, String ip, String userAgent, String params) {
        if (cookieUid != null) {
            var userOpt = userJpa.findById(cookieUid);
            if (userOpt.isPresent()) {
                var user = userOpt.get();
                user.setLastSeenAt(OffsetDateTime.now());
                user.setLastSeenIp(ip);
                userJpa.save(user);
                return new UserResolution(user.getId(), false);
            }
        }

        var user = new UserEntity();
        user.setLastSeenAt(OffsetDateTime.now());
        user.setLastSeenIp(ip);
        user.setFirstVisitUserAgent(userAgent);
        user.setFirstVisitHeaders(params);
        user = userJpa.save(user);
        return new UserResolution(user.getId(), true);
    }

    public record UserResolution(UUID userId, boolean isNew) {}
}
