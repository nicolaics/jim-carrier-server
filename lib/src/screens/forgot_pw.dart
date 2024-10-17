import 'package:flutter/material.dart';
import 'package:jim/src/constants/sizes.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:get/get.dart';
import 'package:jim/src/screens/otp_screen.dart';

class ForgetPassword extends StatefulWidget {
  const ForgetPassword({Key? key}) : super(key: key);
  @override
  _ForgetPassword createState() => _ForgetPassword();
}
class _ForgetPassword extends State<ForgetPassword> {
  final TextEditingController _emailController = TextEditingController();
  Widget build(BuildContext context){
    return Scaffold(
      body: SingleChildScrollView(
      child: Container(
        padding: const EdgeInsets.all(tDefaultSize),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(height: tDefaultSize * 4),
            Text(
              'Forgot Password?',
              style: GoogleFonts.anton(fontSize: 40),
            ),
            Text('Please enter the email you use to sign in with.',
                style: GoogleFonts.cormorant(
                    fontSize: 15, color: Colors.black),
                textAlign: TextAlign.center),
            Form(
                child: Container(
                  padding: EdgeInsets.symmetric(vertical: 20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      TextFormField(
                        controller: _emailController,
                        decoration: InputDecoration(
                            prefixIcon: Icon(Icons.person_outline_outlined),
                            labelText: "Email",
                            hintText: "your@gmail.com",
                            border: OutlineInputBorder()),
                      ),
                      SizedBox(
                        height: 30,
                      ),
                      SizedBox(
                          width: double.infinity,
                          child: ElevatedButton(
                              onPressed: () => Get.to(()=> const OtpScreen()),
                              style: OutlinedButton.styleFrom(
                                  shape: RoundedRectangleBorder(),
                                  backgroundColor: Colors.black),
                              child: Text('SEND CODE',
                                  style: TextStyle(
                                      color: Colors.white, fontSize: 20)))),
                    ],
                  ),
                )),
          ],
        ),
      ),
      )
    );
  }
}