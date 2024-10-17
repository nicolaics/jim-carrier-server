import 'package:flutter/material.dart';
import 'package:jim/src/screens/login_screen.dart';
import 'package:jim/src/screens/register_screen.dart';
import 'package:get/get.dart';
import 'package:jim/src/screens/welcome.dart';


void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return const GetMaterialApp(
      debugShowCheckedModeBanner: false,
      defaultTransition: Transition.leftToRightWithFade,
      transitionDuration: Duration(milliseconds: 500),
      home:  WelcomeScreen(),
    );

  }
}
