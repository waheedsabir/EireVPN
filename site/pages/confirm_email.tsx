import React, { useState, useEffect } from 'react';
import { LayoutMain } from '../components/Layout';
import { useRouter } from 'next/router';
import API from '../service/APIService';
import ErrorMessage from '../components/ErrorMessage';
import SuccessMessage from '../components/SuccessMessage';
import useAsync from '../hooks/useAsync';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Container from 'react-bootstrap/Container';
import Card from 'react-bootstrap/Card';

export default function ConfirmEmailPage(): JSX.Element {
  const router = useRouter();
  const token = router.query.token;
  const { data, loading, error } = useAsync(() => API.ConfirmEmail(token.toString()));

  useEffect(() => {
    if (!token) {
      router.push('/');
    }
  });

  if (loading) {
    return <div></div>;
  }

  return (
    <LayoutMain>
      <Container>
        <Row>
          <Col />
          <Col>
            <Card className="password-reset-card">
              <ErrorMessage show={!!error} error={error} />
              <SuccessMessage show={!error} message="Thank you for confirming your email" />
            </Card>
          </Col>
          <Col />
        </Row>
      </Container>
    </LayoutMain>
  );
}

ConfirmEmailPage.getInitialProps = async () => {
  return {};
};
