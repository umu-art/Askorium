package ru.askorium.core.user.domain;

import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "users")
public class UserEntity extends BaseEntity {

    private OffsetDateTime lastSeenAt;

    private String lastSeenIp;

    private String firstVisitUserAgent;

    private String firstVisitHeaders;

}
